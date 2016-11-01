// Package crypto wraps go's builtin crypto libraries to make common operations easier.
// It also helps reduce the number of crypto/etc imports that must be put in each file.
package crypto

import (
	builtin "crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"io/ioutil"
	"math/big"
)

const (
	Md5           = builtin.MD5
	Sha256        = builtin.SHA256
	Sha256WithRsa = x509.SHA256WithRSA
)

// Redeclare so they don't have to be imported
type Certificate *x509.Certificate
type PrivateKey builtin.PrivateKey
type PublicKey builtin.PublicKey

type PemType string

const (
	PemX509        = PemType("CERTIFICATE")
	PemX509Pair    = PemType("CERTIFICATE PAIR")
	PemX509Trusted = PemType("TRUSTED CERTIFICATE")
	PemCertRequest = PemType("CERTIFICATE REQUEST")
	PemRsaPrivate  = PemType("RSA PRIVATE KEY")
	PemDsaPrivate  = PemType("DSA PRIVATE KEY")
	PemPkcs7       = PemType("PKCS7")
	PemPkcs8       = PemType("ENCRYPTED PRIVATE KEY")
	PemPkcs8Info   = PemType("PRIVATE KEY")
	PemDhParams    = PemType("DH PARAMETERS")
	PemSslParams   = PemType("SSL SESSION PARAMETERS")
	PemDsaParams   = PemType("DSA PARAMETERS")
	PemEcParams    = PemType("EC PARAMETERS")
	PemEcPrivate   = PemType("EC PRIVATE KEY")
	PemPkix        = PemType("PUBLIC KEY")
)

type ECDSASignature struct {
	R, S *big.Int
}

// SignSha256 accepts a message and an ECDSA or RSA private key and returns a
// signature of the digest.
//
// N.B. When using an RSA key, PKCS1 v1.5 is assumed.
func SignSha256(key PrivateKey, msg []byte) (signature []byte, err error) {
	digest := sha256.Sum256(msg)
	switch k := key.(type) {
	case *rsa.PrivateKey:
		return rsa.SignPKCS1v15(rand.Reader, k, Sha256, digest[:])
	case *ecdsa.PrivateKey:
		r, s, err := ecdsa.Sign(rand.Reader, k, digest[:])
		if err != nil {
			return nil, err
		}
		return asn1.Marshal(ECDSASignature{r, s})
	default:
		return nil, &PrivateKeyTypeError{key}
	}
}

// VerifySha256 accepts a message, signature, and ECDSA or RSA public key and
// verifies the message was signed with the corresponding private key.
//
// N.B. When using an RSA key, PKCS1 v1.5 is assumed.
func VerifySha256(pub PublicKey, msg []byte, sig []byte) bool {
	digest := sha256.Sum256(msg)
	switch p := pub.(type) {
	case *rsa.PublicKey:
		return (rsa.VerifyPKCS1v15(p, Sha256, digest[:], sig) == nil)
	case *ecdsa.PublicKey:
		var ec ECDSASignature
		extra, err := asn1.Unmarshal(sig, &ec)
		if err != nil || len(extra) > 0 {
			return false
		}
		return ecdsa.Verify(p, digest[:], ec.R, ec.S)
	default:
		return false
	}
}

// MustGenerateRsaKey wraps rsa.GenerateKey but panics if a key cannot be generated.
// It simplifies key generation in unittests and one-off scripts.
func MustGenerateRsaKey(size int) *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, size)
	if err != nil {
		panic(err)
	}
	return key
}

// LoadCertificate loads an X509 certificate in PEM format.
func LoadCertificate(path string) (Certificate, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data) // ignoring remaining data
	if PemType(block.Type) != PemX509 {
		return nil, &PemTypeError{PemX509, PemType(block.Type)}
	}

	return x509.ParseCertificate(block.Bytes)
}

// LoadCertificateString loads an X509 certificate from a string.
func LoadCertificateString(text string) (Certificate, error) {
	block, _ := pem.Decode([]byte(text)) // ignoring remaining data
	if PemType(block.Type) != PemX509 {
		return nil, &PemTypeError{PemX509, PemType(block.Type)}
	}

	return x509.ParseCertificate(block.Bytes)
}

// LoadCertificateBytes loads an X509 certificate from a byte slice.
func LoadCertificateBytes(data []byte) (Certificate, error) {
	block, _ := pem.Decode(data) // ignoring remaining data
	if PemType(block.Type) != PemX509 {
		return nil, &PemTypeError{PemX509, PemType(block.Type)}
	}

	return x509.ParseCertificate(block.Bytes)
}

// MustLoadCertificate is like LoadCertificate but panics if the key cannot be loaded.
// It simplifies safe intialization of global variables.
func MustLoadCertificate(path string) Certificate {
	cert, err := LoadCertificate(path)
	if err != nil {
		panic(err)
	}
	return cert
}

// LoadPrivateKey loads an RSA or ECDSA private key in PEM format.
// It may be wrapped in unencrypted PKCS8 format, but DES keys are not supported.
func LoadPrivateKey(path string) (PrivateKey, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data) // ignoring remaining data
	switch PemType(block.Type) {
	case PemPkcs8Info:
		return x509.ParsePKCS8PrivateKey(block.Bytes)
	case PemRsaPrivate:
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case PemEcPrivate:
		return x509.ParseECPrivateKey(block.Bytes)
	default:
		return nil, &PemTypeError{"* PRIVATE KEY", PemType(block.Type)}
	}
}

// LoadPrivateKeyString loads an RSA or ECDSA private key from a string.
func LoadPrivateKeyString(text string) (PrivateKey, error) {
	block, _ := pem.Decode([]byte(text)) // ignoring remaining data
	switch PemType(block.Type) {
	case PemPkcs8Info:
		return x509.ParsePKCS8PrivateKey(block.Bytes)
	case PemRsaPrivate:
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case PemEcPrivate:
		return x509.ParseECPrivateKey(block.Bytes)
	default:
		return nil, &PemTypeError{"* PRIVATE KEY", PemType(block.Type)}
	}
}

// LoadPrivateKeyBytes loads an RSA or ECDSA private key from a byte slice.
func LoadPrivateKeyBytes(data []byte) (PrivateKey, error) {
	block, _ := pem.Decode(data) // ignoring remaining data
	switch PemType(block.Type) {
	case PemPkcs8Info:
		return x509.ParsePKCS8PrivateKey(block.Bytes)
	case PemRsaPrivate:
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case PemEcPrivate:
		return x509.ParseECPrivateKey(block.Bytes)
	default:
		return nil, &PemTypeError{"* PRIVATE KEY", PemType(block.Type)}
	}
}

// MustLoadPrivateKey is like LoadPrivateKey but panics if the key cannot be loaded.
// It simplifies safe intialization of global variables.
func MustLoadPrivateKey(path string) PrivateKey {
	key, err := LoadPrivateKey(path)
	if err != nil {
		panic(err)
	}
	return key
}

// LoadPublicKey loads an RSA or ECDSA public key in PEM format.
func LoadPublicKey(path string) (PublicKey, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data) // ignoring remaining data
	return x509.ParsePKIXPublicKey(block.Bytes)
}

// LoadPublicKeyString loads an RSA or ECDSA public key from a string.
func LoadPublicKeyString(text string) (PublicKey, error) {
	block, _ := pem.Decode([]byte(text)) // ignoring remaining data
	return x509.ParsePKIXPublicKey(block.Bytes)
}

// LoadPublicKeyBytes loads an RSA or ECDSA public key from a byte slice.
func LoadPublicKeyBytes(data []byte) (PublicKey, error) {
	block, _ := pem.Decode(data) // ignoring remaining data
	return x509.ParsePKIXPublicKey(block.Bytes)
}

// MustLoadPublicKey is like LoadPublicKey but panics if the key cannot be loaded.
// It simplifies safe intialization of global variables.
func MustLoadPublicKey(path string) PublicKey {
	key, err := LoadPublicKey(path)
	if err != nil {
		panic(err)
	}
	return key
}

type PrivateKeyTypeError struct {
	Key PrivateKey
}

func (err *PrivateKeyTypeError) Error() string {
	return `crypto: unsupported private key type`
}

type PublicKeyTypeError struct {
	Key PublicKey
}

func (err *PublicKeyTypeError) Error() string {
	return `crypto: unsupported public key type`
}

type PemTypeError struct {
	Expected PemType
	Received PemType
}

func (err *PemTypeError) Error() string {
	return `pem: expected "` + string(err.Expected) + `" but recieved "` + string(err.Received) + `"`
}
