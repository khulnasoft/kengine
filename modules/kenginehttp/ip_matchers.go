// Copyright 2015 Matthew Holt and The Kengine Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kenginehttp

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"reflect"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types/ref"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/khulnasoft/kengine/v2"
	"github.com/khulnasoft/kengine/v2/kengineconfig/kenginefile"
	"github.com/khulnasoft/kengine/v2/internal"
)

// MatchRemoteIP matches requests by the remote IP address,
// i.e. the IP address of the direct connection to Kengine.
type MatchRemoteIP struct {
	// The IPs or CIDR ranges to match.
	Ranges []string `json:"ranges,omitempty"`

	// cidrs and zones vars should aligned always in the same
	// length and indexes for matching later
	cidrs  []*netip.Prefix
	zones  []string
	logger *zap.Logger
}

// MatchClientIP matches requests by the client IP address,
// i.e. the resolved address, considering trusted proxies.
type MatchClientIP struct {
	// The IPs or CIDR ranges to match.
	Ranges []string `json:"ranges,omitempty"`

	// cidrs and zones vars should aligned always in the same
	// length and indexes for matching later
	cidrs  []*netip.Prefix
	zones  []string
	logger *zap.Logger
}

func init() {
	kengine.RegisterModule(MatchRemoteIP{})
	kengine.RegisterModule(MatchClientIP{})
}

// KengineModule returns the Kengine module information.
func (MatchRemoteIP) KengineModule() kengine.ModuleInfo {
	return kengine.ModuleInfo{
		ID:  "http.matchers.remote_ip",
		New: func() kengine.Module { return new(MatchRemoteIP) },
	}
}

// UnmarshalKenginefile implements kenginefile.Unmarshaler.
func (m *MatchRemoteIP) UnmarshalKenginefile(d *kenginefile.Dispenser) error {
	// iterate to merge multiple matchers into one
	for d.Next() {
		for d.NextArg() {
			if d.Val() == "forwarded" {
				return d.Err("the 'forwarded' option is no longer supported; use the 'client_ip' matcher instead")
			}
			if d.Val() == "private_ranges" {
				m.Ranges = append(m.Ranges, internal.PrivateRangesCIDR()...)
				continue
			}
			m.Ranges = append(m.Ranges, d.Val())
		}
		if d.NextBlock(0) {
			return d.Err("malformed remote_ip matcher: blocks are not supported")
		}
	}
	return nil
}

// CELLibrary produces options that expose this matcher for use in CEL
// expression matchers.
//
// Example:
//
//	expression remote_ip('192.168.0.0/16', '172.16.0.0/12', '10.0.0.0/8')
func (MatchRemoteIP) CELLibrary(ctx kengine.Context) (cel.Library, error) {
	return CELMatcherImpl(
		// name of the macro, this is the function name that users see when writing expressions.
		"remote_ip",
		// name of the function that the macro will be rewritten to call.
		"remote_ip_match_request_list",
		// internal data type of the MatchPath value.
		[]*cel.Type{cel.ListType(cel.StringType)},
		// function to convert a constant list of strings to a MatchPath instance.
		func(data ref.Val) (RequestMatcher, error) {
			refStringList := reflect.TypeOf([]string{})
			strList, err := data.ConvertToNative(refStringList)
			if err != nil {
				return nil, err
			}

			m := MatchRemoteIP{}

			for _, input := range strList.([]string) {
				if input == "forwarded" {
					return nil, errors.New("the 'forwarded' option is no longer supported; use the 'client_ip' matcher instead")
				}
				m.Ranges = append(m.Ranges, input)
			}

			err = m.Provision(ctx)
			return m, err
		},
	)
}

// Provision parses m's IP ranges, either from IP or CIDR expressions.
func (m *MatchRemoteIP) Provision(ctx kengine.Context) error {
	m.logger = ctx.Logger()
	cidrs, zones, err := provisionCidrsZonesFromRanges(m.Ranges)
	if err != nil {
		return err
	}
	m.cidrs = cidrs
	m.zones = zones

	return nil
}

// Match returns true if r matches m.
func (m MatchRemoteIP) Match(r *http.Request) bool {
	if r.TLS != nil && !r.TLS.HandshakeComplete {
		return false // if handshake is not finished, we infer 0-RTT that has not verified remote IP; could be spoofed
	}
	address := r.RemoteAddr
	clientIP, zoneID, err := parseIPZoneFromString(address)
	if err != nil {
		if c := m.logger.Check(zapcore.ErrorLevel, "getting remote "); c != nil {
			c.Write(zap.Error(err))
		}

		return false
	}
	matches, zoneFilter := matchIPByCidrZones(clientIP, zoneID, m.cidrs, m.zones)
	if !matches && !zoneFilter {
		if c := m.logger.Check(zapcore.DebugLevel, "zone ID from remote IP did not match"); c != nil {
			c.Write(zap.String("zone", zoneID))
		}
	}
	return matches
}

// KengineModule returns the Kengine module information.
func (MatchClientIP) KengineModule() kengine.ModuleInfo {
	return kengine.ModuleInfo{
		ID:  "http.matchers.client_ip",
		New: func() kengine.Module { return new(MatchClientIP) },
	}
}

// UnmarshalKenginefile implements kenginefile.Unmarshaler.
func (m *MatchClientIP) UnmarshalKenginefile(d *kenginefile.Dispenser) error {
	// iterate to merge multiple matchers into one
	for d.Next() {
		for d.NextArg() {
			if d.Val() == "private_ranges" {
				m.Ranges = append(m.Ranges, internal.PrivateRangesCIDR()...)
				continue
			}
			m.Ranges = append(m.Ranges, d.Val())
		}
		if d.NextBlock(0) {
			return d.Err("malformed client_ip matcher: blocks are not supported")
		}
	}
	return nil
}

// CELLibrary produces options that expose this matcher for use in CEL
// expression matchers.
//
// Example:
//
//	expression client_ip('192.168.0.0/16', '172.16.0.0/12', '10.0.0.0/8')
func (MatchClientIP) CELLibrary(ctx kengine.Context) (cel.Library, error) {
	return CELMatcherImpl(
		// name of the macro, this is the function name that users see when writing expressions.
		"client_ip",
		// name of the function that the macro will be rewritten to call.
		"client_ip_match_request_list",
		// internal data type of the MatchPath value.
		[]*cel.Type{cel.ListType(cel.StringType)},
		// function to convert a constant list of strings to a MatchPath instance.
		func(data ref.Val) (RequestMatcher, error) {
			refStringList := reflect.TypeOf([]string{})
			strList, err := data.ConvertToNative(refStringList)
			if err != nil {
				return nil, err
			}

			m := MatchClientIP{
				Ranges: strList.([]string),
			}

			err = m.Provision(ctx)
			return m, err
		},
	)
}

// Provision parses m's IP ranges, either from IP or CIDR expressions.
func (m *MatchClientIP) Provision(ctx kengine.Context) error {
	m.logger = ctx.Logger()
	cidrs, zones, err := provisionCidrsZonesFromRanges(m.Ranges)
	if err != nil {
		return err
	}
	m.cidrs = cidrs
	m.zones = zones
	return nil
}

// Match returns true if r matches m.
func (m MatchClientIP) Match(r *http.Request) bool {
	if r.TLS != nil && !r.TLS.HandshakeComplete {
		return false // if handshake is not finished, we infer 0-RTT that has not verified remote IP; could be spoofed
	}
	address := GetVar(r.Context(), ClientIPVarKey).(string)
	clientIP, zoneID, err := parseIPZoneFromString(address)
	if err != nil {
		m.logger.Error("getting client IP", zap.Error(err))
		return false
	}
	matches, zoneFilter := matchIPByCidrZones(clientIP, zoneID, m.cidrs, m.zones)
	if !matches && !zoneFilter {
		m.logger.Debug("zone ID from client IP did not match", zap.String("zone", zoneID))
	}
	return matches
}

func provisionCidrsZonesFromRanges(ranges []string) ([]*netip.Prefix, []string, error) {
	cidrs := []*netip.Prefix{}
	zones := []string{}
	repl := kengine.NewReplacer()
	for _, str := range ranges {
		str = repl.ReplaceAll(str, "")
		// Exclude the zone_id from the IP
		if strings.Contains(str, "%") {
			split := strings.Split(str, "%")
			str = split[0]
			// write zone identifiers in m.zones for matching later
			zones = append(zones, split[1])
		} else {
			zones = append(zones, "")
		}
		if strings.Contains(str, "/") {
			ipNet, err := netip.ParsePrefix(str)
			if err != nil {
				return nil, nil, fmt.Errorf("parsing CIDR expression '%s': %v", str, err)
			}
			cidrs = append(cidrs, &ipNet)
		} else {
			ipAddr, err := netip.ParseAddr(str)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid IP address: '%s': %v", str, err)
			}
			ipNew := netip.PrefixFrom(ipAddr, ipAddr.BitLen())
			cidrs = append(cidrs, &ipNew)
		}
	}
	return cidrs, zones, nil
}

func parseIPZoneFromString(address string) (netip.Addr, string, error) {
	ipStr, _, err := net.SplitHostPort(address)
	if err != nil {
		ipStr = address // OK; probably didn't have a port
	}

	// Some IPv6-Addresses can contain zone identifiers at the end,
	// which are separated with "%"
	zoneID := ""
	if strings.Contains(ipStr, "%") {
		split := strings.Split(ipStr, "%")
		ipStr = split[0]
		zoneID = split[1]
	}

	ipAddr, err := netip.ParseAddr(ipStr)
	if err != nil {
		return netip.IPv4Unspecified(), "", err
	}

	return ipAddr, zoneID, nil
}

func matchIPByCidrZones(clientIP netip.Addr, zoneID string, cidrs []*netip.Prefix, zones []string) (bool, bool) {
	zoneFilter := true
	for i, ipRange := range cidrs {
		if ipRange.Contains(clientIP) {
			// Check if there are zone filters assigned and if they match.
			if zones[i] == "" || zoneID == zones[i] {
				return true, false
			}
			zoneFilter = false
		}
	}
	return false, zoneFilter
}

// Interface guards
var (
	_ RequestMatcher        = (*MatchRemoteIP)(nil)
	_ kengine.Provisioner     = (*MatchRemoteIP)(nil)
	_ kenginefile.Unmarshaler = (*MatchRemoteIP)(nil)
	_ CELLibraryProducer    = (*MatchRemoteIP)(nil)

	_ RequestMatcher        = (*MatchClientIP)(nil)
	_ kengine.Provisioner     = (*MatchClientIP)(nil)
	_ kenginefile.Unmarshaler = (*MatchClientIP)(nil)
	_ CELLibraryProducer    = (*MatchClientIP)(nil)
)