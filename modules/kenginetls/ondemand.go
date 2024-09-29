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

package kenginetls

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/khulnasoft-lab/certmagic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/khulnasoft/kengine/v2"
	"github.com/khulnasoft/kengine/v2/kengineconfig/kenginefile"
)

func init() {
	kengine.RegisterModule(PermissionByHTTP{})
}

// OnDemandConfig configures on-demand TLS, for obtaining
// needed certificates at handshake-time. Because this
// feature can easily be abused, you should use this to
// establish rate limits and/or an internal endpoint that
// Kengine can "ask" if it should be allowed to manage
// certificates for a given hostname.
type OnDemandConfig struct {
	// DEPRECATED. WILL BE REMOVED SOON. Use 'permission' instead.
	Ask string `json:"ask,omitempty"`

	// REQUIRED. A module that will determine whether a
	// certificate is allowed to be loaded from storage
	// or obtained from an issuer on demand.
	PermissionRaw json.RawMessage `json:"permission,omitempty" kengine:"namespace=tls.permission inline_key=module"`
	permission    OnDemandPermission

	// DEPRECATED. An optional rate limit to throttle
	// the checking of storage and the issuance of
	// certificates from handshakes if not already in
	// storage. WILL BE REMOVED IN A FUTURE RELEASE.
	RateLimit *RateLimit `json:"rate_limit,omitempty"`
}

// DEPRECATED. WILL LIKELY BE REMOVED SOON.
// Instead of using this rate limiter, use a proper tool such as a
// level 3 or 4 firewall and/or a permission module to apply rate limits.
type RateLimit struct {
	// A duration value. Storage may be checked and a certificate may be
	// obtained 'burst' times during this interval.
	Interval kengine.Duration `json:"interval,omitempty"`

	// How many times during an interval storage can be checked or a
	// certificate can be obtained.
	Burst int `json:"burst,omitempty"`
}

// OnDemandPermission is a type that can give permission for
// whether a certificate should be allowed to be obtained or
// loaded from storage on-demand.
// EXPERIMENTAL: This API is experimental and subject to change.
type OnDemandPermission interface {
	// CertificateAllowed returns nil if a certificate for the given
	// name is allowed to be either obtained from an issuer or loaded
	// from storage on-demand.
	//
	// The context passed in has the associated *tls.ClientHelloInfo
	// value available at the certmagic.ClientHelloInfoCtxKey key.
	//
	// In the worst case, this function may be called as frequently
	// as every TLS handshake, so it should return as quick as possible
	// to reduce latency. In the normal case, this function is only
	// called when a certificate is needed that is not already loaded
	// into memory ready to serve.
	CertificateAllowed(ctx context.Context, name string) error
}

// PermissionByHTTP determines permission for a TLS certificate by
// making a request to an HTTP endpoint.
type PermissionByHTTP struct {
	// The endpoint to access. It should be a full URL.
	// A query string parameter "domain" will be added to it,
	// containing the domain (or IP) for the desired certificate,
	// like so: `?domain=example.com`. Generally, this endpoint
	// is not exposed publicly to avoid a minor information leak
	// (which domains are serviced by your application).
	//
	// The endpoint must return a 200 OK status if a certificate
	// is allowed; anything else will cause it to be denied.
	// Redirects are not followed.
	Endpoint string `json:"endpoint"`

	logger   *zap.Logger
	replacer *kengine.Replacer
}

// KengineModule returns the Kengine module information.
func (PermissionByHTTP) KengineModule() kengine.ModuleInfo {
	return kengine.ModuleInfo{
		ID:  "tls.permission.http",
		New: func() kengine.Module { return new(PermissionByHTTP) },
	}
}

// UnmarshalKenginefile implements kenginefile.Unmarshaler.
func (p *PermissionByHTTP) UnmarshalKenginefile(d *kenginefile.Dispenser) error {
	if !d.Next() {
		return nil
	}
	if !d.AllArgs(&p.Endpoint) {
		return d.ArgErr()
	}
	return nil
}

func (p *PermissionByHTTP) Provision(ctx kengine.Context) error {
	p.logger = ctx.Logger()
	p.replacer = kengine.NewReplacer()
	return nil
}

func (p PermissionByHTTP) CertificateAllowed(ctx context.Context, name string) error {
	// run replacer on endpoint URL (for environment variables) -- return errors to prevent surprises (#5036)
	askEndpoint, err := p.replacer.ReplaceOrErr(p.Endpoint, true, true)
	if err != nil {
		return fmt.Errorf("preparing 'ask' endpoint: %v", err)
	}

	askURL, err := url.Parse(askEndpoint)
	if err != nil {
		return fmt.Errorf("parsing ask URL: %v", err)
	}
	qs := askURL.Query()
	qs.Set("domain", name)
	askURL.RawQuery = qs.Encode()
	askURLString := askURL.String()

	var remote string
	if chi, ok := ctx.Value(certmagic.ClientHelloInfoCtxKey).(*tls.ClientHelloInfo); ok && chi != nil {
		remote = chi.Conn.RemoteAddr().String()
	}

	if c := p.logger.Check(zapcore.DebugLevel, "asking permission endpoint"); c != nil {
		c.Write(
			zap.String("remote", remote),
			zap.String("domain", name),
			zap.String("url", askURLString),
		)
	}

	resp, err := onDemandAskClient.Get(askURLString)
	if err != nil {
		return fmt.Errorf("checking %v to determine if certificate for hostname '%s' should be allowed: %v",
			askEndpoint, name, err)
	}
	resp.Body.Close()

	if c := p.logger.Check(zapcore.DebugLevel, "response from permission endpoint"); c != nil {
		c.Write(
			zap.String("remote", remote),
			zap.String("domain", name),
			zap.String("url", askURLString),
			zap.Int("status", resp.StatusCode),
		)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("%s: %w %s - non-2xx status code %d", name, ErrPermissionDenied, askEndpoint, resp.StatusCode)
	}

	return nil
}

// ErrPermissionDenied is an error that should be wrapped or returned when the
// configured permission module does not allow a certificate to be issued,
// to distinguish that from other errors such as connection failure.
var ErrPermissionDenied = errors.New("certificate not allowed by permission module")

// These perpetual values are used for on-demand TLS.
var (
	onDemandRateLimiter = certmagic.NewRateLimiter(0, 0)
	onDemandAskClient   = &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return fmt.Errorf("following http redirects is not allowed")
		},
	}
)

// Interface guards
var (
	_ OnDemandPermission = (*PermissionByHTTP)(nil)
	_ kengine.Provisioner  = (*PermissionByHTTP)(nil)
)