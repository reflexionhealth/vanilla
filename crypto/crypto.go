// Package crypto wraps go's builtin crypto libraries to make common operations easy
// and to reduce the number of crypto/etc imports that must be put in each file
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

func MustGenerateRSAKey(size int) *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, size) // tiny key so test runs fast
	if err != nil {
		panic(err)
	}

	return key
}

func LoadCertificatePEM(path string) (*x509.Certificate, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data) // ignoring remaining data
	if block.Type != "CERTIFICATE" {
		return nil, &PEMTypeError{"CERTIFICATE", block.Type}
	}

	return x509.ParseCertificate(block.Bytes)
}

func MustLoadCertificatePEM(path string) *x509.Certificate {
	cert, err := LoadCertificatePEM(path)
	if err != nil {
		panic(err)
	}

	return cert
}

func LoadRSAPrivateKeyPEM(path string) (*rsa.PrivateKey, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data) // ignoring remaining data
	if block.Type != "RSA PRIVATE KEY" {
		return nil, &PEMTypeError{"RSA PRIVATE KEY", block.Type}
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func MustLoadRSAPrivateKeyPEM(path string) *rsa.PrivateKey {
	key, err := LoadRSAPrivateKeyPEM(path)
	if err != nil {
		panic(err)
	}

	return key
}

type PEMTypeError struct {
	Expected string
	Received string
}

func (err *PEMTypeError) Error() string {
	return `pem: expected "` + err.Expected + `" but recieved "` + err.Received + `"`
}
