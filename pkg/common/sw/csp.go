package sw

import (
	"github.com/pkg/errors"
	"io/ioutil"
)

type CSP struct{}

// Sign signs digest using key k.
// The opts argument should be appropriate for the primitive used.
//
// Note that when a signature of a hash of a larger message is needed,
// the caller is responsible for hashing the larger message and passing
// the hash (as digest).
func (csp *CSP) Sign(k Key, digest []byte, opts SignerOpts) (signature []byte, err error) {
	// Validate arguments
	if k == nil {
		return nil, errors.New("Invalid Key. It must not be nil.")
	}
	if len(digest) == 0 {
		return nil, errors.New("Invalid digest. Cannot be empty.")
	}

	switch k.(type) {
	case *sm2PrivateKey, *sm2PublicKey:
		signer := sm2Signer{}
		signature, err = signer.Sign(k, digest, opts)
	case *ecdsaPrivateKey, *ecdsaPublicKey:
		signer := ecdsaSigner{}
		signature, err = signer.Sign(k, digest, opts)
	}

	if err != nil {
		return nil, errors.Wrapf(err, "Failed signing with opts [%v]", opts)
	}

	return
}

// Verify verifies signature against key k and digest
func (csp *CSP) Verify(k Key, signature, digest []byte, opts SignerOpts) (valid bool, err error) {
	// Validate arguments
	if k == nil {
		return false, errors.New("Invalid Key. It must not be nil.")
	}
	if len(signature) == 0 {
		return false, errors.New("Invalid signature. Cannot be empty.")
	}
	if len(digest) == 0 {
		return false, errors.New("Invalid digest. Cannot be empty.")
	}

	switch k.(type) {
	case *sm2PrivateKey:
		verifier := sm2PrivateKeyVerifier{}
		valid, err = verifier.Verify(k, signature, digest, opts)
	case *sm2PublicKey:
		verifier := sm2PublicKeyKeyVerifier{}
		valid, err = verifier.Verify(k, signature, digest, opts)
	case *ecdsaPublicKey:
		verifier := ecdsaPublicKeyKeyVerifier{}
		valid, err = verifier.Verify(k, signature, digest, opts)
	case *ecdsaPrivateKey:
		verifier := ecdsaPrivateKeyVerifier{}
		valid, err = verifier.Verify(k, signature, digest, opts)
	}
	if err != nil {
		return false, errors.Wrapf(err, "Failed verifing with opts [%v]", opts)
	}

	return
}

type SimpleCSP struct {
	*CSP
	*KeyStore
}

func NewSimpleCSP(keyPath, certPath string) (*SimpleCSP, error) {
	ks, err := newKeyStore(keyPath, certPath)
	if err != nil {
		return nil, err
	}
	return &SimpleCSP{
		CSP:      &CSP{},
		KeyStore: ks,
	}, nil
}

type KeyStore struct {
	PrivateKey Key
	PublicKey  Key
}

func newKeyStore(keyPath, certPath string) (*KeyStore, error) {
	keyStore := &KeyStore{}
	if keyPath != "" {
		key, err := ioutil.ReadFile(keyPath)
		if err != nil {
			return nil, err
		}
		priKey, err := ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}
		keyStore.PrivateKey = priKey
	}
	if certPath != "" {
		cert, err := ioutil.ReadFile(certPath)
		if err != nil {
			return nil, err
		}
		pubKey, err := ParsePublicByCertificate(cert)
		if err != nil {
			return nil, err
		}
		keyStore.PublicKey = pubKey
	}

	return keyStore, nil
}

func (s *SimpleCSP) Sign(digest []byte) (signature []byte, err error) {
	return s.CSP.Sign(s.PrivateKey, digest, nil)
}

func (s *SimpleCSP) Verify(signature, digest []byte) (bool, error) {
	return s.CSP.Verify(s.PublicKey, signature, digest, nil)
}
