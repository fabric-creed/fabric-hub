package grpc

import (
	"context"
	"errors"
	"github.com/fabric-creed/cryptogm/tls"
	"github.com/fabric-creed/cryptogm/x509"
	"github.com/fabric-creed/grpc/credentials"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

var (
	ErrClientHandshakeNotImplemented = errors.New("core/comm: client handshakes are not implemented with serverCreds")
	ErrServerHandshakeNotImplemented = errors.New("core/comm: server handshakes are not implemented with clientCreds")
	ErrOverrideHostnameNotSupported  = errors.New("core/comm: OverrideServerName is not supported")

	// alpnProtoStr are the specified application level protocols for gRPC.
	alpnProtoStr = []string{"h2"}
)

// NewServerTransportCredentials returns a new initialized
// grpc/credentials.TransportCredentials
func NewServerTransportCredentials(
	serverConfig *TLSConfig) credentials.TransportCredentials {
	// NOTE: unlike the default grpc/credentials implementation, we do not
	// clone the tls.Config which allows us to update it dynamically
	serverConfig.config.NextProtos = alpnProtoStr
	// GMTLS support
	if serverConfig.config.GMSupport != nil {
		serverConfig.config.MinVersion = tls.VersionGMSSL
	} else {
		serverConfig.config.MinVersion = tls.VersionTLS12
	}

	return &serverCreds{
		serverConfig: serverConfig.config,
	}
}

// serverCreds is an implementation of grpc/credentials.TransportCredentials.
type serverCreds struct {
	serverConfig *tls.Config
}

type TLSConfig struct {
	config *tls.Config
	lock   sync.RWMutex
}

func NewTLSConfig(config *tls.Config) *TLSConfig {
	return &TLSConfig{
		config: config,
	}
}

func (t *TLSConfig) Config() tls.Config {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if t.config != nil {
		return *t.config.Clone()
	}

	return tls.Config{}
}

func (t *TLSConfig) AddClientRootCA(cert *x509.Certificate) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.config.ClientCAs.AddCert(cert)
}

func (t *TLSConfig) SetClientCAs(certPool *x509.CertPool) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.config.ClientCAs = certPool
}

// ClientHandShake is not implemented for `serverCreds`.
func (sc *serverCreds) ClientHandshake(context.Context,
	string, net.Conn) (net.Conn, credentials.AuthInfo, error) {
	return nil, nil, ErrClientHandshakeNotImplemented
}

// ServerHandshake does the authentication handshake for servers.
func (sc *serverCreds) ServerHandshake(rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	conn := tls.Server(rawConn, sc.serverConfig)
	start := time.Now()
	if err := conn.Handshake(); err != nil {
		logrus.Errorf("remote address %s, Server TLS handshake failed in %s with error %s", conn.RemoteAddr().String(), time.Since(start), err)
		return nil, nil, err
	}
	logrus.Debugf("remote address %s, Server TLS handshake completed in %s", conn.RemoteAddr().String(), time.Since(start))
	return conn, credentials.TLSInfo{State: conn.ConnectionState()}, nil
}

// Info provides the ProtocolInfo of this TransportCredentials.
func (sc *serverCreds) Info() credentials.ProtocolInfo {
	return credentials.ProtocolInfo{
		SecurityProtocol: "tls",
		SecurityVersion:  "1.2",
	}
}

// Clone makes a copy of this TransportCredentials.
func (sc *serverCreds) Clone() credentials.TransportCredentials {
	return NewServerTransportCredentials(NewTLSConfig(sc.serverConfig))
}

// OverrideServerName overrides the server name used to verify the hostname
// on the returned certificates from the server.
func (sc *serverCreds) OverrideServerName(string) error {
	return ErrOverrideHostnameNotSupported
}

type DynamicClientCredentials struct {
	TLSConfig  *tls.Config
	TLSOptions []TLSOption
}

func (dtc *DynamicClientCredentials) latestConfig() *tls.Config {
	tlsConfigCopy := dtc.TLSConfig.Clone()
	for _, tlsOption := range dtc.TLSOptions {
		tlsOption(tlsConfigCopy)
	}
	return tlsConfigCopy
}

func (dtc *DynamicClientCredentials) ClientHandshake(ctx context.Context, authority string, rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	creds := credentials.NewTLS(dtc.latestConfig())
	start := time.Now()
	conn, auth, err := creds.ClientHandshake(ctx, authority, rawConn)
	if err != nil {
		logrus.Errorf("remote address %s, Client TLS handshake failed after %s with error: %s", rawConn.RemoteAddr().String(), time.Since(start), err)
	} else {
		logrus.Debugf("remote address %s,Client TLS handshake completed in %s", rawConn.RemoteAddr().String(), time.Since(start))
	}
	return conn, auth, err
}

func (dtc *DynamicClientCredentials) ServerHandshake(rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	return nil, nil, ErrServerHandshakeNotImplemented
}

func (dtc *DynamicClientCredentials) Info() credentials.ProtocolInfo {
	return credentials.NewTLS(dtc.latestConfig()).Info()
}

func (dtc *DynamicClientCredentials) Clone() credentials.TransportCredentials {
	return credentials.NewTLS(dtc.latestConfig())
}

func (dtc *DynamicClientCredentials) OverrideServerName(name string) error {
	dtc.TLSConfig.ServerName = name
	return nil
}
