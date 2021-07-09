package sw

import (
	"crypto/ecdsa"
	"encoding/pem"
	"errors"
	"github.com/fabric-creed/cryptogm/sm2"
	"github.com/fabric-creed/cryptogm/x509"
)

func ParsePrivateKey(keyPEMBlock []byte) (Key, error) {
	var keyDERBlock *pem.Block
	keyDERBlock, keyPEMBlock = pem.Decode(keyPEMBlock)
	if keyDERBlock == nil {
		return nil, errors.New(" failed to find any PEM data in key input")
	}

	if key, err := x509.ParsePKCS8PrivateKey(keyDERBlock.Bytes); err == nil {
		switch key := key.(type) {
		case *ecdsa.PrivateKey:
			return NewEcdsaPrivateKey(key), nil
		case *sm2.PrivateKey:
			return NewSm2PrivateKey(key), nil
		default:
			return nil, errors.New("found unknown private key type in PKCS#8 wrapping")
		}
	}

	if key, err := x509.ParseECPrivateKey(keyDERBlock.Bytes); err == nil {
		switch key := key.(type) {
		case *ecdsa.PrivateKey:
			return NewEcdsaPrivateKey(key), nil
		case *sm2.PrivateKey:
			return NewSm2PrivateKey(key), nil
		default:
			return nil, errors.New("found unknown private key type in PKCS#8 wrapping")
		}
	}

	return nil, errors.New("failed to parse private key")
}

func ParsePublicByCertificate(data []byte) (Key, error) {
	pemBlock, _ := pem.Decode(data)
	cert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}
	switch key := cert.PublicKey.(type) {
	case *ecdsa.PublicKey:
		return NewEcdsaPublicKey(key), nil
	case *sm2.PublicKey:
		return NewSm2PublicKey(key), nil
	default:
		return nil, errors.New("found unknown public key type")
	}
}
