package grpc

import (
	"github.com/fabric-creed/cryptogm/tls"
	"github.com/fabric-creed/cryptogm/x509"
	"github.com/fabric-creed/grpc/credentials"
	"github.com/sirupsen/logrus"
	"sync"
)

// CredentialSupport type manages credentials used for gRPC client connections
type CredentialSupport struct {
	mutex             sync.RWMutex
	appRootCAsByChain map[string][][]byte
	serverRootCAs     [][]byte
	clientCert        tls.Certificate
}

// NewCredentialSupport creates a CredentialSupport instance.
func NewCredentialSupport(rootCAs ...[]byte) *CredentialSupport {
	return &CredentialSupport{
		appRootCAsByChain: make(map[string][][]byte),
		serverRootCAs:     rootCAs,
	}
}

// SetClientCertificate sets the tls.Certificate to use for gRPC client
// connections
func (cs *CredentialSupport) SetClientCertificate(cert tls.Certificate) {
	cs.mutex.Lock()
	cs.clientCert = cert
	cs.mutex.Unlock()
}

// GetClientCertificate returns the client certificate of the CredentialSupport
func (cs *CredentialSupport) GetClientCertificate() tls.Certificate {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()
	return cs.clientCert
}

func (cs *CredentialSupport) GetCredentials() credentials.TransportCredentials {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	var appRootCAs [][]byte
	appRootCAs = append(appRootCAs, cs.serverRootCAs...)
	for _, appRootCA := range cs.appRootCAsByChain {
		appRootCAs = append(appRootCAs, appRootCA...)
	}

	certPool := x509.NewCertPool()
	for _, appRootCA := range appRootCAs {
		err := AddPemToCertPool(appRootCA, certPool)
		if err != nil {
			logrus.Warnf("Failed adding certificates to peer's client TLS trust pool: %s", err)
		}
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cs.clientCert},
		RootCAs:      certPool,
	})
}

func (cs *CredentialSupport) AppRootCAsByChain() map[string][][]byte {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()
	return cs.appRootCAsByChain
}
