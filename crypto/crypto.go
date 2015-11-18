// crypto wraps go's builtin crypto libraries to make our common operations easy
package crypto

import (
	builtin "crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
)

const (
	MD5    = builtin.MD5
	SHA256 = builtin.SHA256

	SHA256WithRSA = x509.SHA256WithRSA
)

func SignSHA256WithRSA(key *rsa.PrivateKey, data []byte) (signature []byte, err error) {
	sum := sha256.Sum256(data)
	return rsa.SignPKCS1v15(rand.Reader, key, SHA256, sum[:])
}

func MustLoadPEMCertificate(path string) *x509.Certificate {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	block, _ := pem.Decode(data) // ignoring remaining data
	if block.Type != "CERTIFICATE" {
		panic("loaded pem must have type \"CERTIFICATE\"")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(err)
	}

	return cert
}

func MustLoadPEMPrivateKey(path string) *rsa.PrivateKey {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	block, _ := pem.Decode(data) // ignoring remaining data
	if block.Type != "RSA PRIVATE KEY" {
		panic("loaded pem must have type \"RSA PRIVATE KEY\"")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	return key
}

func MustGenerateRSAKey(size int) *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, size) // tiny key so test runs fast
	if err != nil {
		panic(err)
	}

	return key
}
