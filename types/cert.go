package types

import (
	"crypto/ecdsa"
	"crypto/x509"
)

type CertificateBundle struct {
	Cert     *x509.Certificate
	PrivKey  *ecdsa.PrivateKey
	CertPool *x509.CertPool // CA证书池
	CertPEM  []byte         // PEM编码的证书
	KeyPEM   []byte         // PEM编码的私钥
}
