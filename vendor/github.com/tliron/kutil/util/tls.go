package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
)

func CreateTLSConfig(certificate []byte, key []byte) (*tls.Config, error) {
	if certificate, err := tls.X509KeyPair(certificate, key); err == nil {
		return &tls.Config{
			Certificates: []tls.Certificate{certificate},
		}, nil
	} else {
		return nil, err
	}
}

func CreateSelfSignedTLSConfig(organization string, host string) (*tls.Config, error) {
	//if privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader); err == nil {
	if privateKey, err := rsa.GenerateKey(rand.Reader, 2048); err == nil {
		if certificate, err := CreateX509Certificate(organization, host, true, true); err == nil {
			if signedCertificate, err := SignX509Certificate(certificate, privateKey, &privateKey.PublicKey); err == nil {
				return &tls.Config{
					Certificates: []tls.Certificate{
						{
							Certificate: [][]byte{signedCertificate.Raw},
							PrivateKey:  privateKey,
						},
					},
				}, nil
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}
