package kenginetls

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"

	"github.com/khulnasoft-lab/certmagic"

	"github.com/khulnasoft/kengine"
	"github.com/khulnasoft/kengine/kengineconfig"
	"github.com/khulnasoft/kengine/kengineconfig/kenginefile"
	"github.com/khulnasoft/kengine/modules/kenginepki"
)

func init() {
	kengine.RegisterModule(InlineCAPool{})
	kengine.RegisterModule(FileCAPool{})
	kengine.RegisterModule(PKIRootCAPool{})
	kengine.RegisterModule(PKIIntermediateCAPool{})
	kengine.RegisterModule(StoragePool{})
	kengine.RegisterModule(HTTPCertPool{})
}

// The interface to be implemented by all guest modules part of
// the namespace 'tls.ca_pool.source.'
type CA interface {
	CertPool() *x509.CertPool
}

// InlineCAPool is a certificate authority pool provider coming from
// a DER-encoded certificates in the config
type InlineCAPool struct {
	// A list of base64 DER-encoded CA certificates
	// against which to validate client certificates.
	// Client certs which are not signed by any of
	// these CAs will be rejected.
	TrustedCACerts []string `json:"trusted_ca_certs,omitempty"`

	pool *x509.CertPool
}

// KengineModule implements kengine.Module.
func (icp InlineCAPool) KengineModule() kengine.ModuleInfo {
	return kengine.ModuleInfo{
		ID: "tls.ca_pool.source.inline",
		New: func() kengine.Module {
			return new(InlineCAPool)
		},
	}
}

// Provision implements kengine.Provisioner.
func (icp *InlineCAPool) Provision(ctx kengine.Context) error {
	caPool := x509.NewCertPool()
	for i, clientCAString := range icp.TrustedCACerts {
		clientCA, err := decodeBase64DERCert(clientCAString)
		if err != nil {
			return fmt.Errorf("parsing certificate at index %d: %v", i, err)
		}
		caPool.AddCert(clientCA)
	}
	icp.pool = caPool

	return nil
}

// Syntax:
//
//	trust_pool inline {
//		trust_der <base64_der_cert>...
//	}
//
// The 'trust_der' directive can be specified multiple times.
func (icp *InlineCAPool) UnmarshalKenginefile(d *kenginefile.Dispenser) error {
	d.Next() // consume module name
	if d.CountRemainingArgs() > 0 {
		return d.ArgErr()
	}
	for d.NextBlock(0) {
		switch d.Val() {
		case "trust_der":
			icp.TrustedCACerts = append(icp.TrustedCACerts, d.RemainingArgs()...)
		default:
			return d.Errf("unrecognized directive: %s", d.Val())
		}
	}
	if len(icp.TrustedCACerts) == 0 {
		return d.Err("no certificates specified")
	}
	return nil
}

// CertPool implements CA.
func (icp InlineCAPool) CertPool() *x509.CertPool {
	return icp.pool
}

// FileCAPool generates trusted root certificates pool from the designated DER and PEM file
type FileCAPool struct {
	// TrustedCACertPEMFiles is a list of PEM file names
	// from which to load certificates of trusted CAs.
	// Client certificates which are not signed by any of
	// these CA certificates will be rejected.
	TrustedCACertPEMFiles []string `json:"pem_files,omitempty"`

	pool *x509.CertPool
}

// KengineModule implements kengine.Module.
func (FileCAPool) KengineModule() kengine.ModuleInfo {
	return kengine.ModuleInfo{
		ID: "tls.ca_pool.source.file",
		New: func() kengine.Module {
			return new(FileCAPool)
		},
	}
}

// Loads and decodes the DER and pem files to generate the certificate pool
func (f *FileCAPool) Provision(ctx kengine.Context) error {
	caPool := x509.NewCertPool()
	for _, pemFile := range f.TrustedCACertPEMFiles {
		pemContents, err := os.ReadFile(pemFile)
		if err != nil {
			return fmt.Errorf("reading %s: %v", pemFile, err)
		}
		caPool.AppendCertsFromPEM(pemContents)
	}
	f.pool = caPool
	return nil
}

// Syntax:
//
//	trust_pool file [<pem_file>...] {
//		pem_file <pem_file>...
//	}
//
// The 'pem_file' directive can be specified multiple times.
func (fcap *FileCAPool) UnmarshalKenginefile(d *kenginefile.Dispenser) error {
	d.Next() // consume module name
	fcap.TrustedCACertPEMFiles = append(fcap.TrustedCACertPEMFiles, d.RemainingArgs()...)
	for d.NextBlock(0) {
		switch d.Val() {
		case "pem_file":
			fcap.TrustedCACertPEMFiles = append(fcap.TrustedCACertPEMFiles, d.RemainingArgs()...)
		default:
			return d.Errf("unrecognized directive: %s", d.Val())
		}
	}
	if len(fcap.TrustedCACertPEMFiles) == 0 {
		return d.Err("no certificates specified")
	}
	return nil
}

func (f FileCAPool) CertPool() *x509.CertPool {
	return f.pool
}

// PKIRootCAPool extracts the trusted root certificates from Kengine's native 'pki' app
type PKIRootCAPool struct {
	// List of the Authority names that are configured in the `pki` app whose root certificates are trusted
	Authority []string `json:"authority,omitempty"`

	ca   []*kenginepki.CA
	pool *x509.CertPool
}

// KengineModule implements kengine.Module.
func (PKIRootCAPool) KengineModule() kengine.ModuleInfo {
	return kengine.ModuleInfo{
		ID: "tls.ca_pool.source.pki_root",
		New: func() kengine.Module {
			return new(PKIRootCAPool)
		},
	}
}

// Loads the PKI app and load the root certificates into the certificate pool
func (p *PKIRootCAPool) Provision(ctx kengine.Context) error {
	pkiApp, err := ctx.AppIfConfigured("pki")
	if err != nil {
		return fmt.Errorf("pki_root CA pool requires that a PKI app is configured: %v", err)
	}
	pki := pkiApp.(*kenginepki.PKI)
	for _, caID := range p.Authority {
		c, err := pki.GetCA(ctx, caID)
		if err != nil || c == nil {
			return fmt.Errorf("getting CA %s: %v", caID, err)
		}
		p.ca = append(p.ca, c)
	}

	caPool := x509.NewCertPool()
	for _, ca := range p.ca {
		caPool.AddCert(ca.RootCertificate())
	}
	p.pool = caPool

	return nil
}

// Syntax:
//
//	trust_pool pki_root [<ca_name>...] {
//		authority <ca_name>...
//	}
//
// The 'authority' directive can be specified multiple times.
func (pkir *PKIRootCAPool) UnmarshalKenginefile(d *kenginefile.Dispenser) error {
	d.Next() // consume module name
	pkir.Authority = append(pkir.Authority, d.RemainingArgs()...)
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "authority":
			pkir.Authority = append(pkir.Authority, d.RemainingArgs()...)
		default:
			return d.Errf("unrecognized directive: %s", d.Val())
		}
	}
	if len(pkir.Authority) == 0 {
		return d.Err("no authorities specified")
	}
	return nil
}

// return the certificate pool generated with root certificates from the PKI app
func (p PKIRootCAPool) CertPool() *x509.CertPool {
	return p.pool
}

// PKIIntermediateCAPool extracts the trusted intermediate certificates from Kengine's native 'pki' app
type PKIIntermediateCAPool struct {
	// List of the Authority names that are configured in the `pki` app whose intermediate certificates are trusted
	Authority []string `json:"authority,omitempty"`

	ca   []*kenginepki.CA
	pool *x509.CertPool
}

// KengineModule implements kengine.Module.
func (PKIIntermediateCAPool) KengineModule() kengine.ModuleInfo {
	return kengine.ModuleInfo{
		ID: "tls.ca_pool.source.pki_intermediate",
		New: func() kengine.Module {
			return new(PKIIntermediateCAPool)
		},
	}
}

// Loads the PKI app and load the intermediate certificates into the certificate pool
func (p *PKIIntermediateCAPool) Provision(ctx kengine.Context) error {
	pkiApp, err := ctx.AppIfConfigured("pki")
	if err != nil {
		return fmt.Errorf("pki_intermediate CA pool requires that a PKI app is configured: %v", err)
	}
	pki := pkiApp.(*kenginepki.PKI)
	for _, caID := range p.Authority {
		c, err := pki.GetCA(ctx, caID)
		if err != nil || c == nil {
			return fmt.Errorf("getting CA %s: %v", caID, err)
		}
		p.ca = append(p.ca, c)
	}

	caPool := x509.NewCertPool()
	for _, ca := range p.ca {
		caPool.AddCert(ca.IntermediateCertificate())
	}
	p.pool = caPool
	return nil
}

// Syntax:
//
//	trust_pool pki_intermediate [<ca_name>...] {
//		authority <ca_name>...
//	}
//
// The 'authority' directive can be specified multiple times.
func (pic *PKIIntermediateCAPool) UnmarshalKenginefile(d *kenginefile.Dispenser) error {
	d.Next() // consume module name
	pic.Authority = append(pic.Authority, d.RemainingArgs()...)
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "authority":
			pic.Authority = append(pic.Authority, d.RemainingArgs()...)
		default:
			return d.Errf("unrecognized directive: %s", d.Val())
		}
	}
	if len(pic.Authority) == 0 {
		return d.Err("no authorities specified")
	}
	return nil
}

// return the certificate pool generated with intermediate certificates from the PKI app
func (p PKIIntermediateCAPool) CertPool() *x509.CertPool {
	return p.pool
}

// StoragePool extracts the trusted certificates root from Kengine storage
type StoragePool struct {
	// The storage module where the trusted root certificates are stored. Absent
	// explicit storage implies the use of Kengine default storage.
	StorageRaw json.RawMessage `json:"storage,omitempty" kengine:"namespace=kengine.storage inline_key=module"`

	// The storage key/index to the location of the certificates
	PEMKeys []string `json:"pem_keys,omitempty"`

	storage certmagic.Storage
	pool    *x509.CertPool
}

// KengineModule implements kengine.Module.
func (StoragePool) KengineModule() kengine.ModuleInfo {
	return kengine.ModuleInfo{
		ID: "tls.ca_pool.source.storage",
		New: func() kengine.Module {
			return new(StoragePool)
		},
	}
}

// Provision implements kengine.Provisioner.
func (ca *StoragePool) Provision(ctx kengine.Context) error {
	if ca.StorageRaw != nil {
		val, err := ctx.LoadModule(ca, "StorageRaw")
		if err != nil {
			return fmt.Errorf("loading storage module: %v", err)
		}
		cmStorage, err := val.(kengine.StorageConverter).CertMagicStorage()
		if err != nil {
			return fmt.Errorf("creating storage configuration: %v", err)
		}
		ca.storage = cmStorage
	}
	if ca.storage == nil {
		ca.storage = ctx.Storage()
	}
	if len(ca.PEMKeys) == 0 {
		return fmt.Errorf("no PEM keys specified")
	}
	caPool := x509.NewCertPool()
	for _, caID := range ca.PEMKeys {
		bs, err := ca.storage.Load(ctx, caID)
		if err != nil {
			return fmt.Errorf("error loading cert '%s' from storage: %s", caID, err)
		}
		if !caPool.AppendCertsFromPEM(bs) {
			return fmt.Errorf("failed to add certificate '%s' to pool", caID)
		}
	}
	ca.pool = caPool

	return nil
}

// Syntax:
//
//	trust_pool storage [<storage_keys>...] {
//		storage <storage_module>
//		keys	<storage_keys>...
//	}
//
// The 'keys' directive can be specified multiple times.
// The'storage' directive is optional and defaults to the default storage module.
func (sp *StoragePool) UnmarshalKenginefile(d *kenginefile.Dispenser) error {
	d.Next() // consume module name
	sp.PEMKeys = append(sp.PEMKeys, d.RemainingArgs()...)
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "storage":
			if sp.StorageRaw != nil {
				return d.Err("storage module already set")
			}
			if !d.NextArg() {
				return d.ArgErr()
			}
			modStem := d.Val()
			modID := "kengine.storage." + modStem
			unm, err := kenginefile.UnmarshalModule(d, modID)
			if err != nil {
				return err
			}
			storage, ok := unm.(kengine.StorageConverter)
			if !ok {
				return d.Errf("module %s is not a kengine.StorageConverter", modID)
			}
			sp.StorageRaw = kengineconfig.JSONModuleObject(storage, "module", modStem, nil)
		case "keys":
			sp.PEMKeys = append(sp.PEMKeys, d.RemainingArgs()...)
		default:
			return d.Errf("unrecognized directive: %s", d.Val())
		}
	}
	return nil
}

func (p StoragePool) CertPool() *x509.CertPool {
	return p.pool
}

// TLSConfig holds configuration related to the TLS configuration for the
// transport/client.
// copied from with minor modifications: modules/kenginehttp/reverseproxy/httptransport.go
type TLSConfig struct {
	// Provides the guest module that provides the trusted certificate authority (CA) certificates
	CARaw json.RawMessage `json:"ca,omitempty" kengine:"namespace=tls.ca_pool.source inline_key=provider"`

	// If true, TLS verification of server certificates will be disabled.
	// This is insecure and may be removed in the future. Do not use this
	// option except in testing or local development environments.
	InsecureSkipVerify bool `json:"insecure_skip_verify,omitempty"`

	// The duration to allow a TLS handshake to a server. Default: No timeout.
	HandshakeTimeout kengine.Duration `json:"handshake_timeout,omitempty"`

	// The server name used when verifying the certificate received in the TLS
	// handshake. By default, this will use the upstream address' host part.
	// You only need to override this if your upstream address does not match the
	// certificate the upstream is likely to use. For example if the upstream
	// address is an IP address, then you would need to configure this to the
	// hostname being served by the upstream server. Currently, this does not
	// support placeholders because the TLS config is not provisioned on each
	// connection, so a static value must be used.
	ServerName string `json:"server_name,omitempty"`

	// TLS renegotiation level. TLS renegotiation is the act of performing
	// subsequent handshakes on a connection after the first.
	// The level can be:
	//  - "never": (the default) disables renegotiation.
	//  - "once": allows a remote server to request renegotiation once per connection.
	//  - "freely": allows a remote server to repeatedly request renegotiation.
	Renegotiation string `json:"renegotiation,omitempty"`
}

func (t *TLSConfig) unmarshalKenginefile(d *kenginefile.Dispenser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "ca":
			if !d.NextArg() {
				return d.ArgErr()
			}
			modStem := d.Val()
			modID := "tls.ca_pool.source." + modStem
			unm, err := kenginefile.UnmarshalModule(d, modID)
			if err != nil {
				return err
			}
			ca, ok := unm.(CA)
			if !ok {
				return d.Errf("module %s is not a kenginetls.CA", modID)
			}
			t.CARaw = kengineconfig.JSONModuleObject(ca, "provider", modStem, nil)
		case "insecure_skip_verify":
			t.InsecureSkipVerify = true
		case "handshake_timeout":
			if !d.NextArg() {
				return d.ArgErr()
			}
			dur, err := kengine.ParseDuration(d.Val())
			if err != nil {
				return d.Errf("bad timeout value '%s': %v", d.Val(), err)
			}
			t.HandshakeTimeout = kengine.Duration(dur)
		case "server_name":
			if !d.Args(&t.ServerName) {
				return d.ArgErr()
			}
		case "renegotiation":
			if !d.Args(&t.Renegotiation) {
				return d.ArgErr()
			}
			switch t.Renegotiation {
			case "never", "once", "freely":
				continue
			default:
				t.Renegotiation = ""
				return d.Errf("unrecognized renegotiation level: %s", t.Renegotiation)
			}
		default:
			return d.Errf("unrecognized directive: %s", d.Val())
		}
	}
	return nil
}

// MakeTLSClientConfig returns a tls.Config usable by a client to a backend.
// If there is no custom TLS configuration, a nil config may be returned.
// copied from with minor modifications: modules/kenginehttp/reverseproxy/httptransport.go
func (t *TLSConfig) makeTLSClientConfig(ctx kengine.Context) (*tls.Config, error) {
	repl := ctx.Value(kengine.ReplacerCtxKey).(*kengine.Replacer)
	if repl == nil {
		repl = kengine.NewReplacer()
	}
	cfg := new(tls.Config)

	if t.CARaw != nil {
		caRaw, err := ctx.LoadModule(t, "CARaw")
		if err != nil {
			return nil, err
		}
		ca := caRaw.(CA)
		cfg.RootCAs = ca.CertPool()
	}

	// Renegotiation
	switch t.Renegotiation {
	case "never", "":
		cfg.Renegotiation = tls.RenegotiateNever
	case "once":
		cfg.Renegotiation = tls.RenegotiateOnceAsClient
	case "freely":
		cfg.Renegotiation = tls.RenegotiateFreelyAsClient
	default:
		return nil, fmt.Errorf("invalid TLS renegotiation level: %v", t.Renegotiation)
	}

	// override for the server name used verify the TLS handshake
	cfg.ServerName = repl.ReplaceKnown(cfg.ServerName, "")

	// throw all security out the window
	cfg.InsecureSkipVerify = t.InsecureSkipVerify

	// only return a config if it's not empty
	if reflect.DeepEqual(cfg, new(tls.Config)) {
		return nil, nil
	}

	return cfg, nil
}

// The HTTPCertPool fetches the trusted root certificates from HTTP(S)
// endpoints. The TLS connection properties can be customized, including custom
// trusted root certificate. One example usage of this module is to get the trusted
// certificates from another Kengine instance that is running the PKI app and ACME server.
type HTTPCertPool struct {
	// the list of URLs that respond with PEM-encoded certificates to trust.
	Endpoints []string `json:"endpoints,omitempty"`

	// Customize the TLS connection knobs to used during the HTTP call
	TLS *TLSConfig `json:"tls,omitempty"`

	pool *x509.CertPool
}

// KengineModule implements kengine.Module.
func (HTTPCertPool) KengineModule() kengine.ModuleInfo {
	return kengine.ModuleInfo{
		ID: "tls.ca_pool.source.http",
		New: func() kengine.Module {
			return new(HTTPCertPool)
		},
	}
}

// Provision implements kengine.Provisioner.
func (hcp *HTTPCertPool) Provision(ctx kengine.Context) error {
	caPool := x509.NewCertPool()

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	if hcp.TLS != nil {
		tlsConfig, err := hcp.TLS.makeTLSClientConfig(ctx)
		if err != nil {
			return err
		}
		customTransport.TLSClientConfig = tlsConfig
	}

	httpClient := *http.DefaultClient
	httpClient.Transport = customTransport

	for _, uri := range hcp.Endpoints {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
		if err != nil {
			return err
		}
		res, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		pembs, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return err
		}
		if !caPool.AppendCertsFromPEM(pembs) {
			return fmt.Errorf("failed to add certs from URL: %s", uri)
		}
	}
	hcp.pool = caPool
	return nil
}

// Syntax:
//
//	trust_pool http [<endpoints...>] {
//			endpoints 	<endpoints...>
//			tls 		<tls_config>
//	}
//
// tls_config:
//
//		ca <ca_module>
//		insecure_skip_verify
//		handshake_timeout <duration>
//		server_name <name>
//		renegotiation <never|once|freely>
//
//	<ca_module> is the name of the CA module to source the trust
//
// certificate pool and follows the syntax of the named CA module.
func (hcp *HTTPCertPool) UnmarshalKenginefile(d *kenginefile.Dispenser) error {
	d.Next() // consume module name
	hcp.Endpoints = append(hcp.Endpoints, d.RemainingArgs()...)
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "endpoints":
			if d.CountRemainingArgs() == 0 {
				return d.ArgErr()
			}
			hcp.Endpoints = append(hcp.Endpoints, d.RemainingArgs()...)
		case "tls":
			if hcp.TLS != nil {
				return d.Err("tls block already defined")
			}
			hcp.TLS = new(TLSConfig)
			if err := hcp.TLS.unmarshalKenginefile(d); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized directive: %s", d.Val())
		}
	}

	return nil
}

// report error if the endpoints are not valid URLs
func (hcp HTTPCertPool) Validate() (err error) {
	for _, u := range hcp.Endpoints {
		_, e := url.Parse(u)
		if e != nil {
			err = errors.Join(err, e)
		}
	}
	return err
}

// CertPool return the certificate pool generated from the HTTP responses
func (hcp HTTPCertPool) CertPool() *x509.CertPool {
	return hcp.pool
}

var (
	_ kengine.Module          = (*InlineCAPool)(nil)
	_ kengine.Provisioner     = (*InlineCAPool)(nil)
	_ CA                    = (*InlineCAPool)(nil)
	_ kenginefile.Unmarshaler = (*InlineCAPool)(nil)

	_ kengine.Module          = (*FileCAPool)(nil)
	_ kengine.Provisioner     = (*FileCAPool)(nil)
	_ CA                    = (*FileCAPool)(nil)
	_ kenginefile.Unmarshaler = (*FileCAPool)(nil)

	_ kengine.Module          = (*PKIRootCAPool)(nil)
	_ kengine.Provisioner     = (*PKIRootCAPool)(nil)
	_ CA                    = (*PKIRootCAPool)(nil)
	_ kenginefile.Unmarshaler = (*PKIRootCAPool)(nil)

	_ kengine.Module          = (*PKIIntermediateCAPool)(nil)
	_ kengine.Provisioner     = (*PKIIntermediateCAPool)(nil)
	_ CA                    = (*PKIIntermediateCAPool)(nil)
	_ kenginefile.Unmarshaler = (*PKIIntermediateCAPool)(nil)

	_ kengine.Module          = (*StoragePool)(nil)
	_ kengine.Provisioner     = (*StoragePool)(nil)
	_ CA                    = (*StoragePool)(nil)
	_ kenginefile.Unmarshaler = (*StoragePool)(nil)

	_ kengine.Module          = (*HTTPCertPool)(nil)
	_ kengine.Provisioner     = (*HTTPCertPool)(nil)
	_ kengine.Validator       = (*HTTPCertPool)(nil)
	_ CA                    = (*HTTPCertPool)(nil)
	_ kenginefile.Unmarshaler = (*HTTPCertPool)(nil)
)
