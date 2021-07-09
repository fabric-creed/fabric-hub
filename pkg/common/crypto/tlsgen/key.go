package tlsgen

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"

	"github.com/fabric-creed/cryptogm/sm2"
	"github.com/fabric-creed/cryptogm/x509"
	"github.com/pkg/errors"
)

func newPrivateKey(isGM bool) (pri interface{}, pub interface{}, privateBytes []byte, err error) {
	if isGM {
		pri, err = sm2.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, nil, err
		}
		pub = pri.(*sm2.PrivateKey).Public()
	} else {
		pri, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, nil, nil, err
		}
		pub = pri.(*ecdsa.PrivateKey).Public()
	}

	privateBytes, err = x509.MarshalECPrivateKey(pri)
	if err != nil {
		return nil, nil, nil, err
	}
	return pri, pub, privateBytes, nil
}

func newCertTemplate() (x509.Certificate, error) {
	sn, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return x509.Certificate{}, err
	}
	return x509.Certificate{
		Subject:      pkix.Name{SerialNumber: sn.String()},
		NotBefore:    time.Now().Add(time.Hour * (-24)),
		NotAfter:     time.Now().Add(time.Hour * 24),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		SerialNumber: sn,
	}, nil
}

func newCertKeyPair(isCA bool, isServer bool, host string, certSigner crypto.Signer, parent *x509.Certificate, isGM bool) (*CertKeyPair, error) {
	privateKey, publicKey, privateBytes, err := newPrivateKey(isGM)
	if err != nil {
		return nil, err
	}
	var signer crypto.Signer
	switch privateKey.(type) {
	case *sm2.PrivateKey:
		signer = privateKey.(*sm2.PrivateKey)
	case *ecdsa.PrivateKey:
		signer = privateKey.(*ecdsa.PrivateKey)
	}
	var pub crypto.PublicKey
	switch publicKey.(type) {
	case *sm2.PublicKey:
		pub = publicKey.(*sm2.PublicKey)
	case *ecdsa.PrivateKey:
		pub = publicKey.(*ecdsa.PublicKey)
	}

	template, err := newCertTemplate()
	if err != nil {
		return nil, err
	}

	tenYearsFromNow := time.Now().Add(time.Hour * 24 * 365 * 10)
	if isCA {
		template.NotAfter = tenYearsFromNow
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign | x509.KeyUsageCRLSign
		template.ExtKeyUsage = []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		}
		template.BasicConstraintsValid = true
	} else {
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	}
	if isServer {
		template.NotAfter = tenYearsFromNow
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
		if ip := net.ParseIP(host); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, host)
		}
	}
	// If no parent cert, it's a self signed cert
	if parent == nil || certSigner == nil {
		parent = &template
		certSigner = signer
	}
	rawBytes, err := x509.CreateCertificate(rand.Reader, &template, parent, pub, certSigner)
	if err != nil {
		return nil, err
	}
	pubKey := encodePEM("CERTIFICATE", rawBytes)

	block, _ := pem.Decode(pubKey)
	if block == nil { // Never comes unless x509 or pem has bug
		return nil, errors.Errorf("%s: wrong PEM encoding", pubKey)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	privKey := encodePEM("EC PRIVATE KEY", privateBytes)
	return &CertKeyPair{
		Key:     privKey,
		Cert:    pubKey,
		Signer:  signer,
		TLSCert: cert,
	}, nil
}

func encodePEM(keyType string, data []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: keyType, Bytes: data})
}
