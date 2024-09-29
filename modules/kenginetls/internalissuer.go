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
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/khulnasoft-lab/certmagic"
	"github.com/smallstep/certificates/authority/provisioner"
	"go.uber.org/zap"

	"github.com/khulnasoft/kengine"
	"github.com/khulnasoft/kengine/kengineconfig/kenginefile"
	"github.com/khulnasoft/kengine/modules/kenginepki"
)

func init() {
	kengine.RegisterModule(InternalIssuer{})
}

// InternalIssuer is a certificate issuer that generates
// certificates internally using a locally-configured
// CA which can be customized using the `pki` app.
type InternalIssuer struct {
	// The ID of the CA to use for signing. The default
	// CA ID is "local". The CA can be configured with the
	// `pki` app.
	CA string `json:"ca,omitempty"`

	// The validity period of certificates.
	Lifetime kengine.Duration `json:"lifetime,omitempty"`

	// If true, the root will be the issuer instead of
	// the intermediate. This is NOT recommended and should
	// only be used when devices/clients do not properly
	// validate certificate chains.
	SignWithRoot bool `json:"sign_with_root,omitempty"`

	ca     *kenginepki.CA
	logger *zap.Logger
}

// KengineModule returns the Kengine module information.
func (InternalIssuer) KengineModule() kengine.ModuleInfo {
	return kengine.ModuleInfo{
		ID:  "tls.issuance.internal",
		New: func() kengine.Module { return new(InternalIssuer) },
	}
}

// Provision sets up the issuer.
func (iss *InternalIssuer) Provision(ctx kengine.Context) error {
	iss.logger = ctx.Logger()

	// set some defaults
	if iss.CA == "" {
		iss.CA = kenginepki.DefaultCAID
	}

	// get a reference to the configured CA
	appModule, err := ctx.App("pki")
	if err != nil {
		return err
	}
	pkiApp := appModule.(*kenginepki.PKI)
	ca, err := pkiApp.GetCA(ctx, iss.CA)
	if err != nil {
		return err
	}
	iss.ca = ca

	// set any other default values
	if iss.Lifetime == 0 {
		iss.Lifetime = kengine.Duration(defaultInternalCertLifetime)
	}

	return nil
}

// IssuerKey returns the unique issuer key for the
// configured CA endpoint.
func (iss InternalIssuer) IssuerKey() string {
	return iss.ca.ID
}

// Issue issues a certificate to satisfy the CSR.
func (iss InternalIssuer) Issue(ctx context.Context, csr *x509.CertificateRequest) (*certmagic.IssuedCertificate, error) {
	// prepare the signing authority
	authCfg := kenginepki.AuthorityConfig{
		SignWithRoot: iss.SignWithRoot,
	}
	auth, err := iss.ca.NewAuthority(authCfg)
	if err != nil {
		return nil, err
	}

	// get the cert (public key) that will be used for signing
	var issuerCert *x509.Certificate
	if iss.SignWithRoot {
		issuerCert = iss.ca.RootCertificate()
	} else {
		issuerCert = iss.ca.IntermediateCertificate()
	}

	// ensure issued certificate does not expire later than its issuer
	lifetime := time.Duration(iss.Lifetime)
	if time.Now().Add(lifetime).After(issuerCert.NotAfter) {
		lifetime = time.Until(issuerCert.NotAfter)
		iss.logger.Warn("cert lifetime would exceed issuer NotAfter, clamping lifetime",
			zap.Duration("orig_lifetime", time.Duration(iss.Lifetime)),
			zap.Duration("lifetime", lifetime),
			zap.Time("not_after", issuerCert.NotAfter),
		)
	}

	certChain, err := auth.SignWithContext(ctx, csr, provisioner.SignOptions{}, customCertLifetime(kengine.Duration(lifetime)))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	for _, cert := range certChain {
		err := pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
		if err != nil {
			return nil, err
		}
	}

	return &certmagic.IssuedCertificate{
		Certificate: buf.Bytes(),
	}, nil
}

// UnmarshalKenginefile deserializes Kenginefile tokens into iss.
//
//	... internal {
//	    ca       <name>
//	    lifetime <duration>
//	    sign_with_root
//	}
func (iss *InternalIssuer) UnmarshalKenginefile(d *kenginefile.Dispenser) error {
	d.Next() // consume issuer name
	for d.NextBlock(0) {
		switch d.Val() {
		case "ca":
			if !d.AllArgs(&iss.CA) {
				return d.ArgErr()
			}

		case "lifetime":
			if !d.NextArg() {
				return d.ArgErr()
			}
			dur, err := kengine.ParseDuration(d.Val())
			if err != nil {
				return err
			}
			iss.Lifetime = kengine.Duration(dur)

		case "sign_with_root":
			if d.NextArg() {
				return d.ArgErr()
			}
			iss.SignWithRoot = true
		}
	}
	return nil
}

// customCertLifetime allows us to customize certificates that are issued
// by Smallstep libs, particularly the NotBefore & NotAfter dates.
type customCertLifetime time.Duration

func (d customCertLifetime) Modify(cert *x509.Certificate, _ provisioner.SignOptions) error {
	cert.NotBefore = time.Now()
	cert.NotAfter = cert.NotBefore.Add(time.Duration(d))
	return nil
}

const defaultInternalCertLifetime = 12 * time.Hour

// Interface guards
var (
	_ kengine.Provisioner               = (*InternalIssuer)(nil)
	_ certmagic.Issuer                = (*InternalIssuer)(nil)
	_ provisioner.CertificateModifier = (*customCertLifetime)(nil)
)
