package util

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"
)

func ParseX509Certificates(bytes []byte) ([]*x509.Certificate, error) {
	var certificates []*x509.Certificate

	for len(bytes) > 0 {
		var block *pem.Block
		block, bytes = pem.Decode(bytes)
		if block != nil {
			if certificate, err := x509.ParseCertificate(block.Bytes); err == nil {
				certificates = append(certificates, certificate)
			} else {
				return nil, err
			}
		} else {
			bytes = nil
		}
	}

	return certificates, nil
}

func ParseX509CertPool(bytes []byte) (*x509.CertPool, error) {
	if certificates, err := ParseX509Certificates(bytes); err == nil {
		if len(certificates) > 0 {
			certPool := x509.NewCertPool()
			for _, certificate := range certificates {
				certPool.AddCert(certificate)
			}
			return certPool, nil
		} else {
			return nil, nil
		}
	} else {
		return nil, err
	}
}

var serialNumberLimit = new(big.Int).Lsh(big.NewInt(1), 128)

func CreateX509Certificate(organization string, host string, rsa bool, ca bool) (*x509.Certificate, error) {
	// See: https://golang.org/src/crypto/tls/generate_cert.go

	if serialNumber, err := rand.Int(rand.Reader, serialNumberLimit); err == nil {
		now := time.Now()
		certificate := x509.Certificate{
			Subject: pkix.Name{
				Organization: []string{organization},
			},
			DNSNames:     []string{host},
			SerialNumber: serialNumber,
			KeyUsage:     x509.KeyUsageDigitalSignature,
			ExtKeyUsage: []x509.ExtKeyUsage{
				x509.ExtKeyUsageServerAuth,
			},
			//SignatureAlgorithm:    x509.SHA256WithRSA,
			BasicConstraintsValid: true,
			NotBefore:             now,
			NotAfter:              now.Add(365 * 24 * time.Hour), // one year
		}
		if rsa {
			certificate.KeyUsage |= x509.KeyUsageKeyEncipherment
		}
		if ca {
			certificate.IsCA = true
			certificate.KeyUsage |= x509.KeyUsageCertSign
		}
		return &certificate, nil
	} else {
		return nil, err
	}
}

func SignX509Certificate(certificate *x509.Certificate, privateKey any, publicKey any) (*x509.Certificate, error) {
	if certificateBytes, err := x509.CreateCertificate(rand.Reader, certificate, certificate, publicKey, privateKey); err == nil {
		return x509.ParseCertificate(certificateBytes)
	} else {
		return nil, err
	}
}
