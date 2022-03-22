// Package crypto contains various cryptography-related convenience functions.
package crypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"strings"
)

// LoadKeyAndCertificate reads certificate(s) and a private key in PKCS #1 or #8
// form from files in PEM format.
func LoadKeyAndCertificate(keyPath string, certificatePath string) (*tls.Certificate, error) {
	privateKey, err := LoadKey(keyPath)
	if err != nil {
		return nil, err
	}

	certificate, err := LoadCertificate(certificatePath)
	if err != nil {
		return nil, err
	}

	certificate.PrivateKey = privateKey

	return certificate, nil
}

// LoadKey reads a private key in PKCS #1 or #8 form from a file in PEM format.
func LoadKey(path string) (crypto.PrivateKey, error) {
	rawData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	for {
		block, rest := pem.Decode(rawData)
		if block == nil {
			break
		}

		if strings.HasSuffix(block.Type, "PRIVATE KEY") {
			if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
				return key, nil
			}

			if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
				switch key := key.(type) {
				case *rsa.PrivateKey, *ecdsa.PrivateKey:
					return key, nil
				default:
					return nil, errors.New("unknown key type in PKCS#8 wrapping, unable to load key")
				}
			}

			if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
				return key, nil
			}
		}

		rawData = rest
	}

	return nil, errors.New("no private key found, unable to load key")
}

// LoadCertificate reads certificate(s) from a file in PEM format.
func LoadCertificate(path string) (*tls.Certificate, error) {
	rawData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var certificate tls.Certificate

	for {
		block, rest := pem.Decode(rawData)
		if block == nil {
			break
		}

		if block.Type == "CERTIFICATE" {
			certificate.Certificate = append(certificate.Certificate, block.Bytes)
		}

		rawData = rest
	}

	if len(certificate.Certificate) == 0 {
		return nil, errors.New("no certificate found, unable to load certificates")
	}

	return &certificate, nil
}

// LoadX509Certificate reads certificate(s) from a file in PEM format.
func LoadX509Certificate(path string) ([]*x509.Certificate, error) {
	rawData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var certificates []*x509.Certificate

	for {
		var block *pem.Block
		block, rawData = pem.Decode(rawData)
		if block == nil {
			break
		}

		if block.Type != "CERTIFICATE" {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}
		certificates = append(certificates, cert)
	}

	if len(certificates) == 0 {
		return nil, errors.New("no certificate found, unable to load certificates")
	}

	return certificates, nil
}
