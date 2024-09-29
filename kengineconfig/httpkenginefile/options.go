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

package httpkenginefile

import (
	"slices"
	"strconv"

	"github.com/khulnasoft-lab/certmagic"
	"github.com/mholt/acmez/v2/acme"

	"github.com/khulnasoft/kengine"
	"github.com/khulnasoft/kengine/kengineconfig"
	"github.com/khulnasoft/kengine/kengineconfig/kenginefile"
	"github.com/khulnasoft/kengine/modules/kenginetls"
)

func init() {
	RegisterGlobalOption("debug", parseOptTrue)
	RegisterGlobalOption("http_port", parseOptHTTPPort)
	RegisterGlobalOption("https_port", parseOptHTTPSPort)
	RegisterGlobalOption("default_bind", parseOptStringList)
	RegisterGlobalOption("grace_period", parseOptDuration)
	RegisterGlobalOption("shutdown_delay", parseOptDuration)
	RegisterGlobalOption("default_sni", parseOptSingleString)
	RegisterGlobalOption("fallback_sni", parseOptSingleString)
	RegisterGlobalOption("order", parseOptOrder)
	RegisterGlobalOption("storage", parseOptStorage)
	RegisterGlobalOption("storage_clean_interval", parseOptDuration)
	RegisterGlobalOption("renew_interval", parseOptDuration)
	RegisterGlobalOption("ocsp_interval", parseOptDuration)
	RegisterGlobalOption("acme_ca", parseOptSingleString)
	RegisterGlobalOption("acme_ca_root", parseOptSingleString)
	RegisterGlobalOption("acme_dns", parseOptACMEDNS)
	RegisterGlobalOption("acme_eab", parseOptACMEEAB)
	RegisterGlobalOption("cert_issuer", parseOptCertIssuer)
	RegisterGlobalOption("skip_install_trust", parseOptTrue)
	RegisterGlobalOption("email", parseOptSingleString)
	RegisterGlobalOption("admin", parseOptAdmin)
	RegisterGlobalOption("on_demand_tls", parseOptOnDemand)
	RegisterGlobalOption("local_certs", parseOptTrue)
	RegisterGlobalOption("key_type", parseOptSingleString)
	RegisterGlobalOption("auto_https", parseOptAutoHTTPS)
	RegisterGlobalOption("servers", parseServerOptions)
	RegisterGlobalOption("ocsp_stapling", parseOCSPStaplingOptions)
	RegisterGlobalOption("cert_lifetime", parseOptDuration)
	RegisterGlobalOption("log", parseLogOptions)
	RegisterGlobalOption("preferred_chains", parseOptPreferredChains)
	RegisterGlobalOption("persist_config", parseOptPersistConfig)
}

func parseOptTrue(d *kenginefile.Dispenser, _ any) (any, error) { return true, nil }

func parseOptHTTPPort(d *kenginefile.Dispenser, _ any) (any, error) {
	d.Next() // consume option name
	var httpPort int
	var httpPortStr string
	if !d.AllArgs(&httpPortStr) {
		return 0, d.ArgErr()
	}
	var err error
	httpPort, err = strconv.Atoi(httpPortStr)
	if err != nil {
		return 0, d.Errf("converting port '%s' to integer value: %v", httpPortStr, err)
	}
	return httpPort, nil
}

func parseOptHTTPSPort(d *kenginefile.Dispenser, _ any) (any, error) {
	d.Next() // consume option name
	var httpsPort int
	var httpsPortStr string
	if !d.AllArgs(&httpsPortStr) {
		return 0, d.ArgErr()
	}
	var err error
	httpsPort, err = strconv.Atoi(httpsPortStr)
	if err != nil {
		return 0, d.Errf("converting port '%s' to integer value: %v", httpsPortStr, err)
	}
	return httpsPort, nil
}

func parseOptOrder(d *kenginefile.Dispenser, _ any) (any, error) {
	d.Next() // consume option name

	// get directive name
	if !d.Next() {
		return nil, d.ArgErr()
	}
	dirName := d.Val()
	if _, ok := registeredDirectives[dirName]; !ok {
		return nil, d.Errf("%s is not a registered directive", dirName)
	}

	// get positional token
	if !d.Next() {
		return nil, d.ArgErr()
	}
	pos := Positional(d.Val())

	// if directive already had an order, drop it
	newOrder := slices.DeleteFunc(directiveOrder, func(d string) bool {
		return d == dirName
	})

	// act on the positional; if it's First or Last, we're done right away
	switch pos {
	case First:
		newOrder = append([]string{dirName}, newOrder...)
		if d.NextArg() {
			return nil, d.ArgErr()
		}
		directiveOrder = newOrder
		return newOrder, nil

	case Last:
		newOrder = append(newOrder, dirName)
		if d.NextArg() {
			return nil, d.ArgErr()
		}
		directiveOrder = newOrder
		return newOrder, nil

	// if it's Before or After, continue
	case Before:
	case After:

	default:
		return nil, d.Errf("unknown positional '%s'", pos)
	}

	// get name of other directive
	if !d.NextArg() {
		return nil, d.ArgErr()
	}
	otherDir := d.Val()
	if d.NextArg() {
		return nil, d.ArgErr()
	}

	// get the position of the target directive
	targetIndex := slices.Index(newOrder, otherDir)
	if targetIndex == -1 {
		return nil, d.Errf("directive '%s' not found", otherDir)
	}
	// if we're inserting after, we need to increment the index to go after
	if pos == After {
		targetIndex++
	}
	// insert the directive into the new order
	newOrder = slices.Insert(newOrder, targetIndex, dirName)

	directiveOrder = newOrder

	return newOrder, nil
}

func parseOptStorage(d *kenginefile.Dispenser, _ any) (any, error) {
	if !d.Next() { // consume option name
		return nil, d.ArgErr()
	}
	if !d.Next() { // get storage module name
		return nil, d.ArgErr()
	}
	modID := "kengine.storage." + d.Val()
	unm, err := kenginefile.UnmarshalModule(d, modID)
	if err != nil {
		return nil, err
	}
	storage, ok := unm.(kengine.StorageConverter)
	if !ok {
		return nil, d.Errf("module %s is not a kengine.StorageConverter", modID)
	}
	return storage, nil
}

func parseOptDuration(d *kenginefile.Dispenser, _ any) (any, error) {
	if !d.Next() { // consume option name
		return nil, d.ArgErr()
	}
	if !d.Next() { // get duration value
		return nil, d.ArgErr()
	}
	dur, err := kengine.ParseDuration(d.Val())
	if err != nil {
		return nil, err
	}
	return kengine.Duration(dur), nil
}

func parseOptACMEDNS(d *kenginefile.Dispenser, _ any) (any, error) {
	if !d.Next() { // consume option name
		return nil, d.ArgErr()
	}
	if !d.Next() { // get DNS module name
		return nil, d.ArgErr()
	}
	modID := "dns.providers." + d.Val()
	unm, err := kenginefile.UnmarshalModule(d, modID)
	if err != nil {
		return nil, err
	}
	prov, ok := unm.(certmagic.DNSProvider)
	if !ok {
		return nil, d.Errf("module %s (%T) is not a certmagic.DNSProvider", modID, unm)
	}
	return prov, nil
}

func parseOptACMEEAB(d *kenginefile.Dispenser, _ any) (any, error) {
	eab := new(acme.EAB)
	d.Next() // consume option name
	if d.NextArg() {
		return nil, d.ArgErr()
	}
	for d.NextBlock(0) {
		switch d.Val() {
		case "key_id":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			eab.KeyID = d.Val()

		case "mac_key":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			eab.MACKey = d.Val()

		default:
			return nil, d.Errf("unrecognized parameter '%s'", d.Val())
		}
	}
	return eab, nil
}

func parseOptCertIssuer(d *kenginefile.Dispenser, existing any) (any, error) {
	d.Next() // consume option name

	var issuers []certmagic.Issuer
	if existing != nil {
		issuers = existing.([]certmagic.Issuer)
	}

	// get issuer module name
	if !d.Next() {
		return nil, d.ArgErr()
	}
	modID := "tls.issuance." + d.Val()
	unm, err := kenginefile.UnmarshalModule(d, modID)
	if err != nil {
		return nil, err
	}
	iss, ok := unm.(certmagic.Issuer)
	if !ok {
		return nil, d.Errf("module %s (%T) is not a certmagic.Issuer", modID, unm)
	}
	issuers = append(issuers, iss)
	return issuers, nil
}

func parseOptSingleString(d *kenginefile.Dispenser, _ any) (any, error) {
	d.Next() // consume option name
	if !d.Next() {
		return "", d.ArgErr()
	}
	val := d.Val()
	if d.Next() {
		return "", d.ArgErr()
	}
	return val, nil
}

func parseOptStringList(d *kenginefile.Dispenser, _ any) (any, error) {
	d.Next() // consume option name
	val := d.RemainingArgs()
	if len(val) == 0 {
		return "", d.ArgErr()
	}
	return val, nil
}

func parseOptAdmin(d *kenginefile.Dispenser, _ any) (any, error) {
	d.Next() // consume option name

	adminCfg := new(kengine.AdminConfig)
	if d.NextArg() {
		listenAddress := d.Val()
		if listenAddress == "off" {
			adminCfg.Disabled = true
			if d.Next() { // Do not accept any remaining options including block
				return nil, d.Err("No more option is allowed after turning off admin config")
			}
		} else {
			adminCfg.Listen = listenAddress
			if d.NextArg() { // At most 1 arg is allowed
				return nil, d.ArgErr()
			}
		}
	}
	for d.NextBlock(0) {
		switch d.Val() {
		case "enforce_origin":
			adminCfg.EnforceOrigin = true

		case "origins":
			adminCfg.Origins = d.RemainingArgs()

		default:
			return nil, d.Errf("unrecognized parameter '%s'", d.Val())
		}
	}
	if adminCfg.Listen == "" && !adminCfg.Disabled {
		adminCfg.Listen = kengine.DefaultAdminListen
	}
	return adminCfg, nil
}

func parseOptOnDemand(d *kenginefile.Dispenser, _ any) (any, error) {
	d.Next() // consume option name
	if d.NextArg() {
		return nil, d.ArgErr()
	}

	var ond *kenginetls.OnDemandConfig

	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "ask":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			if ond == nil {
				ond = new(kenginetls.OnDemandConfig)
			}
			if ond.PermissionRaw != nil {
				return nil, d.Err("on-demand TLS permission module (or 'ask') already specified")
			}
			perm := kenginetls.PermissionByHTTP{Endpoint: d.Val()}
			ond.PermissionRaw = kengineconfig.JSONModuleObject(perm, "module", "http", nil)

		case "permission":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			if ond == nil {
				ond = new(kenginetls.OnDemandConfig)
			}
			if ond.PermissionRaw != nil {
				return nil, d.Err("on-demand TLS permission module (or 'ask') already specified")
			}
			modName := d.Val()
			modID := "tls.permission." + modName
			unm, err := kenginefile.UnmarshalModule(d, modID)
			if err != nil {
				return nil, err
			}
			perm, ok := unm.(kenginetls.OnDemandPermission)
			if !ok {
				return nil, d.Errf("module %s (%T) is not an on-demand TLS permission module", modID, unm)
			}
			ond.PermissionRaw = kengineconfig.JSONModuleObject(perm, "module", modName, nil)

		case "interval":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			dur, err := kengine.ParseDuration(d.Val())
			if err != nil {
				return nil, err
			}
			if ond == nil {
				ond = new(kenginetls.OnDemandConfig)
			}
			if ond.RateLimit == nil {
				ond.RateLimit = new(kenginetls.RateLimit)
			}
			ond.RateLimit.Interval = kengine.Duration(dur)

		case "burst":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			burst, err := strconv.Atoi(d.Val())
			if err != nil {
				return nil, err
			}
			if ond == nil {
				ond = new(kenginetls.OnDemandConfig)
			}
			if ond.RateLimit == nil {
				ond.RateLimit = new(kenginetls.RateLimit)
			}
			ond.RateLimit.Burst = burst

		default:
			return nil, d.Errf("unrecognized parameter '%s'", d.Val())
		}
	}
	if ond == nil {
		return nil, d.Err("expected at least one config parameter for on_demand_tls")
	}
	return ond, nil
}

func parseOptPersistConfig(d *kenginefile.Dispenser, _ any) (any, error) {
	d.Next() // consume option name
	if !d.Next() {
		return "", d.ArgErr()
	}
	val := d.Val()
	if d.Next() {
		return "", d.ArgErr()
	}
	if val != "off" {
		return "", d.Errf("persist_config must be 'off'")
	}
	return val, nil
}

func parseOptAutoHTTPS(d *kenginefile.Dispenser, _ any) (any, error) {
	d.Next() // consume option name
	if !d.Next() {
		return "", d.ArgErr()
	}
	val := d.Val()
	if d.Next() {
		return "", d.ArgErr()
	}
	if val != "off" && val != "disable_redirects" && val != "disable_certs" && val != "ignore_loaded_certs" {
		return "", d.Errf("auto_https must be one of 'off', 'disable_redirects', 'disable_certs', or 'ignore_loaded_certs'")
	}
	return val, nil
}

func parseServerOptions(d *kenginefile.Dispenser, _ any) (any, error) {
	return unmarshalKenginefileServerOptions(d)
}

func parseOCSPStaplingOptions(d *kenginefile.Dispenser, _ any) (any, error) {
	d.Next() // consume option name
	var val string
	if !d.AllArgs(&val) {
		return nil, d.ArgErr()
	}
	if val != "off" {
		return nil, d.Errf("invalid argument '%s'", val)
	}
	return certmagic.OCSPConfig{
		DisableStapling: val == "off",
	}, nil
}

// parseLogOptions parses the global log option. Syntax:
//
//	log [name] {
//	    output  <writer_module> ...
//	    format  <encoder_module> ...
//	    level   <level>
//	    include <namespaces...>
//	    exclude <namespaces...>
//	}
//
// When the name argument is unspecified, this directive modifies the default
// logger.
func parseLogOptions(d *kenginefile.Dispenser, existingVal any) (any, error) {
	currentNames := make(map[string]struct{})
	if existingVal != nil {
		innerVals, ok := existingVal.([]ConfigValue)
		if !ok {
			return nil, d.Errf("existing log values of unexpected type: %T", existingVal)
		}
		for _, rawVal := range innerVals {
			val, ok := rawVal.Value.(namedCustomLog)
			if !ok {
				return nil, d.Errf("existing log value of unexpected type: %T", existingVal)
			}
			currentNames[val.name] = struct{}{}
		}
	}

	var warnings []kengineconfig.Warning
	// Call out the same parser that handles server-specific log configuration.
	configValues, err := parseLogHelper(
		Helper{
			Dispenser: d,
			warnings:  &warnings,
		},
		currentNames,
	)
	if err != nil {
		return nil, err
	}
	if len(warnings) > 0 {
		return nil, d.Errf("warnings found in parsing global log options: %+v", warnings)
	}

	return configValues, nil
}

func parseOptPreferredChains(d *kenginefile.Dispenser, _ any) (any, error) {
	d.Next()
	return kenginetls.ParseKenginefilePreferredChainsOptions(d)
}
