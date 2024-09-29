package kenginetls

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/klauspost/cpuid"
)

// supportedCipherSuites is the unordered map of cipher suite
// string names to their definition in crypto/tls.
// TODO: might not be needed much longer, see:
// https://github.com/golang/go/issues/30325
var supportedCipherSuites = map[string]uint16{
	"ECDHE_ECDSA_AES256_GCM_SHA384":      tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	"ECDHE_RSA_AES256_GCM_SHA384":        tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	"ECDHE_ECDSA_AES128_GCM_SHA256":      tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	"ECDHE_RSA_AES128_GCM_SHA256":        tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	"ECDHE_ECDSA_WITH_CHACHA20_POLY1305": tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	"ECDHE_RSA_WITH_CHACHA20_POLY1305":   tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
	"ECDHE_RSA_AES256_CBC_SHA":           tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	"ECDHE_RSA_AES128_CBC_SHA":           tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	"ECDHE_ECDSA_AES256_CBC_SHA":         tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	"ECDHE_ECDSA_AES128_CBC_SHA":         tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	"RSA_AES256_CBC_SHA":                 tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	"RSA_AES128_CBC_SHA":                 tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	"ECDHE_RSA_3DES_EDE_CBC_SHA":         tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
	"RSA_3DES_EDE_CBC_SHA":               tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
}

// defaultCipherSuites is the ordered list of all the cipher
// suites we want to support by default, assuming AES-NI
// (hardware acceleration for AES).
var defaultCipherSuitesWithAESNI = []uint16{
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
}

// defaultCipherSuites is the ordered list of all the cipher
// suites we want to support by default, assuming lack of
// AES-NI (NO hardware acceleration for AES).
var defaultCipherSuitesWithoutAESNI = []uint16{
	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
}

// getOptimalDefaultCipherSuites returns an appropriate cipher
// suite to use depending on the hardware support for AES.
//
// See https://github.com/mholt/kengine/issues/1674
func getOptimalDefaultCipherSuites() []uint16 {
	if cpuid.CPU.AesNi() {
		return defaultCipherSuitesWithAESNI
	}
	return defaultCipherSuitesWithoutAESNI
}

// supportedCurves is the unordered map of supported curves.
// https://golang.org/pkg/crypto/tls/#CurveID
var supportedCurves = map[string]tls.CurveID{
	"X25519": tls.X25519,
	"P256":   tls.CurveP256,
	"P384":   tls.CurveP384,
	"P521":   tls.CurveP521,
}

// defaultCurves is the list of only the curves we want to use
// by default, in descending order of preference.
//
// This list should only include curves which are fast by design
// (e.g. X25519) and those for which an optimized assembly
// implementation exists (e.g. P256). The latter ones can be
// found here:
// https://github.com/golang/go/tree/master/src/crypto/elliptic
var defaultCurves = []tls.CurveID{
	tls.X25519,
	tls.CurveP256,
}

// supportedProtocols is a map of supported protocols.
// HTTP/2 only supports TLS 1.2 and higher.
var supportedProtocols = map[string]uint16{
	"tls1.0": tls.VersionTLS10,
	"tls1.1": tls.VersionTLS11,
	"tls1.2": tls.VersionTLS12,
	"tls1.3": tls.VersionTLS13,
}

// publicKeyAlgorithms is the map of supported public key algorithms.
var publicKeyAlgorithms = map[string]x509.PublicKeyAlgorithm{
	"rsa":   x509.RSA,
	"dsa":   x509.DSA,
	"ecdsa": x509.ECDSA,
}