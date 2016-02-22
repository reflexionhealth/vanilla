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
)

// SignSha256 accepts a message and an ECDSA or RSA private key and returns a
// signature of the digest.
//
// N.B. When using an RSA key, PKCS1 v1.5 signatures are preferred over PSS,
// because PSS is still doesn't seem widely supported/tested in the wild (Feb 2016),
// and additionally there are no known defects of PKCS1 v1.5.
// To sign with PSS, import the crypto/rsa and use rsa.SignPSS/VerifyPSS.
func SignSha256(key builtin.PrivateKey, msg []byte) (signature []byte, err error) {
	digest := sha256.Sum256(msg)
	switch k := key.(type) {
	case *rsa.PrivateKey:
		return rsa.SignPKCS1v15(rand.Reader, k, Sha256, digest[:])
	case *ecdsa.PrivateKey:
		r, s, err := ecdsa.Sign(rand.Reader, k, digest[:])
		if err != nil {
			return nil, err
		}
		return asn1.Marshal(ecdsaSignature{r, s})
	default:
		return nil, &KeyTypeError{key}
	}
}

type ecdsaSignature struct {
	R, S *big.Int
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
func LoadCertificate(path string) (*x509.Certificate, error) {
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

// MustLoadCertificate is like LoadCertificate but panics if the key cannot be loaded.
// It simplifies safe intialization of global variables.
func MustLoadCertificate(path string) *x509.Certificate {
	cert, err := LoadCertificate(path)
	if err != nil {
		panic(err)
	}
	return cert
}

// LoadPrivateKey loads an RSA or ECDSA private key in PEM format.
// It may be wrapped in unencrypted PKCS8 format, but DES keys are not supported.
func LoadPrivateKey(path string) (builtin.PrivateKey, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data) // ignoring remaining data
	switch PemType(block.Type) {
	case PemPkcs8:
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
func MustLoadPrivateKey(path string) builtin.PrivateKey {
	key, err := LoadPrivateKey(path)
	if err != nil {
		panic(err)
	}
	return key
}

type KeyTypeError struct {
	Key builtin.PrivateKey
}

func (err *KeyTypeError) Error() string {
	return `crypto: unsupported private key type"`
}

type PemTypeError struct {
	Expected PemType
	Received PemType
}

func (err *PemTypeError) Error() string {
	return `pem: expected "` + string(err.Expected) + `" but recieved "` + string(err.Received) + `"`
}
