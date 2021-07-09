package grpc

import (
	"github.com/fabric-creed/cryptogm/tls"
	"github.com/fabric-creed/cryptogm/x509"
	"github.com/fabric-creed/fabric-hub/pkg/common/crypto/tlsgen"
	"github.com/fabric-creed/grpc"
	"github.com/fabric-creed/grpc/keepalive"
	"github.com/pkg/errors"
	"io/ioutil"
	"time"
)

// Configuration defaults
var (
	// Max send and receive bytes for grpc clients and servers
	MaxRecvMsgSize = 100 * 1024 * 1024
	MaxSendMsgSize = 100 * 1024 * 1024
	// Default peer keepalive options
	DefaultKeepaliveOptions = KeepaliveOptions{
		ClientInterval:    time.Duration(1) * time.Minute,  // 1 min
		ClientTimeout:     time.Duration(20) * time.Second, // 20 sec - gRPC default
		ServerInterval:    time.Duration(2) * time.Hour,    // 2 hours - gRPC default
		ServerTimeout:     time.Duration(20) * time.Second, // 20 sec - gRPC default
		ServerMinInterval: time.Duration(1) * time.Minute,  // match ClientInterval
	}
	// strong TLS cipher suites
	DefaultTLSCipherSuites = []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	}
	// strong GM TLS cipher suites
	DefaultGMTLSCipherSuites = []uint16{
		tls.GMTLS_SM2_WITH_SM4_SM3,
		tls.GMTLS_ECDHE_SM2_WITH_SM4_SM3,
	}
	// default connection timeout
	DefaultConnectionTimeout = 10 * time.Second
)

// ServerConfig defines the parameters for configuring a GRPCServer instance
type ServerConfig struct {
	// ConnectionTimeout specifies the timeout for connection establishment
	// for all new connections
	ConnectionTimeout time.Duration
	// SecOpts defines the security parameters
	SecOpts SecureOptions
	// KaOpts defines the keepalive parameters
	KaOpts KeepaliveOptions
	// StreamInterceptors specifies a list of interceptors to apply to
	// streaming RPCs.  They are executed in order.
	StreamInterceptors []grpc.StreamServerInterceptor
	// UnaryInterceptors specifies a list of interceptors to apply to unary
	// RPCs.  They are executed in order.
	UnaryInterceptors []grpc.UnaryServerInterceptor
	// HealthCheckEnabled enables the gRPC Health Checking Protocol for the server
	HealthCheckEnabled bool
}

// ClientConfig defines the parameters for configuring a GRPCClient instance
type ClientConfig struct {
	// SecOpts defines the security parameters
	SecOpts SecureOptions
	// KaOpts defines the keepalive parameters
	KaOpts KeepaliveOptions
	// Timeout specifies how long the client will block when attempting to
	// establish a connection
	Timeout time.Duration
	// AsyncConnect makes connection creation non blocking
	AsyncConnect bool
}

// Clone clones this ClientConfig
func (cc ClientConfig) Clone() ClientConfig {
	shallowClone := cc
	return shallowClone
}

// SecureOptions defines the security parameters (e.g. TLS) for a
// GRPCServer or GRPCClient instance
type SecureOptions struct {
	// VerifyCertificate, if not nil, is called after normal
	// certificate verification by either a TLS client or server.
	// If it returns a non-nil error, the handshake is aborted and that error results.
	VerifyCertificate func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error
	// PEM-encoded X509 public key to be used for TLS communication
	Certificate []byte
	// PEM-encoded private key to be used for TLS communication
	Key []byte
	// Set of PEM-encoded X509 certificate authorities used by clients to
	// verify server certificates
	ServerRootCAs [][]byte
	// Set of PEM-encoded X509 certificate authorities used by servers to
	// verify client certificates
	ClientRootCAs [][]byte
	// Whether or not to use TLS for communication
	UseTLS bool
	// Whether or not TLS client must present certificates for authentication
	RequireClientCert bool
	// CipherSuites is a list of supported cipher suites for TLS
	CipherSuites []uint16
	// TimeShift makes TLS handshakes time sampling shift to the past by a given duration
	TimeShift time.Duration
}

// KeepaliveOptions is used to set the gRPC keepalive settings for both
// clients and servers
type KeepaliveOptions struct {
	// ClientInterval is the duration after which if the client does not see
	// any activity from the server it pings the server to see if it is alive
	ClientInterval time.Duration
	// ClientTimeout is the duration the client waits for a response
	// from the server after sending a ping before closing the connection
	ClientTimeout time.Duration
	// ServerInterval is the duration after which if the server does not see
	// any activity from the client it pings the client to see if it is alive
	ServerInterval time.Duration
	// ServerTimeout is the duration the server waits for a response
	// from the client after sending a ping before closing the connection
	ServerTimeout time.Duration
	// ServerMinInterval is the minimum permitted time between client pings.
	// If clients send pings more frequently, the server will disconnect them
	ServerMinInterval time.Duration
}

// ServerKeepaliveOptions returns gRPC keepalive options for server.
func ServerKeepaliveOptions(ka KeepaliveOptions) []grpc.ServerOption {
	var serverOpts []grpc.ServerOption
	kap := keepalive.ServerParameters{
		Time:    ka.ServerInterval,
		Timeout: ka.ServerTimeout,
	}
	serverOpts = append(serverOpts, grpc.KeepaliveParams(kap))
	kep := keepalive.EnforcementPolicy{
		MinTime: ka.ServerMinInterval,
		// allow keepalive w/o rpc
		PermitWithoutStream: true,
	}
	serverOpts = append(serverOpts, grpc.KeepaliveEnforcementPolicy(kep))
	return serverOpts
}

// ClientKeepaliveOptions returns gRPC keepalive options for clients.
func ClientKeepaliveOptions(ka KeepaliveOptions) []grpc.DialOption {
	var dialOpts []grpc.DialOption
	kap := keepalive.ClientParameters{
		Time:                ka.ClientInterval,
		Timeout:             ka.ClientTimeout,
		PermitWithoutStream: true,
	}
	dialOpts = append(dialOpts, grpc.WithKeepaliveParams(kap))
	return dialOpts
}

type genTLSCertFunc func() (*tlsgen.CertKeyPair, error)

// Config defines configuration of a Client
type Config struct {
	CertPath       string
	KeyPath        string
	PeerCACertPath string
	Timeout        time.Duration
}

// ToSecureOptions converts this Config to SecureOptions.
// The given function generates a self signed client TLS certificate if
// the TLS certificate and key aren't present at the config
func ClientSecureOptions(certPath, keyPath, caCertPath, serverCACertPath string, isGm bool) (SecureOptions, error) {
	if serverCACertPath == "" {
		return SecureOptions{}, nil
	}
	serverCABytes, err := loadFile(serverCACertPath)
	if err != nil {
		return SecureOptions{}, errors.WithStack(err)
	}
	var keyBytes, certBytes, caCertBytes []byte
	// If TLS key and certificate aren't given, generate a self signed one on the fly
	if keyPath == "" && certPath == "" {
		tlsCert, err := newSelfSignedTLSCert(isGm)
		if err != nil {
			return SecureOptions{}, err
		}
		keyBytes, certBytes = tlsCert.Key, tlsCert.Cert
	} else {
		keyBytes, err = loadFile(keyPath)
		if err != nil {
			return SecureOptions{}, errors.WithStack(err)
		}
		certBytes, err = loadFile(certPath)
		if err != nil {
			return SecureOptions{}, errors.WithStack(err)
		}
		caCertBytes, err = loadFile(caCertPath)
		if err != nil {
			return SecureOptions{}, errors.WithStack(err)
		}
	}
	return SecureOptions{
		Key:               keyBytes,
		Certificate:       certBytes,
		UseTLS:            true,
		ServerRootCAs:     [][]byte{serverCABytes},
		ClientRootCAs:     [][]byte{caCertBytes},
		RequireClientCert: true,
	}, nil
}

func ServerSecureOptions(useTLS bool, certPath, keyPath, caCertPath string, requireClientCert bool, clientCACertsPath []string) (SecureOptions, error) {
	so := SecureOptions{
		UseTLS:            false,
		RequireClientCert: false,
		CipherSuites:      DefaultTLSCipherSuites,
	}
	if useTLS {
		so.UseTLS = useTLS
		caBytes, err := loadFile(caCertPath)
		if err != nil {
			return SecureOptions{}, errors.WithStack(err)
		}
		so.ServerRootCAs = append(so.ServerRootCAs, caBytes)
		keyBytes, err := loadFile(keyPath)
		if err != nil {
			return SecureOptions{}, errors.WithStack(err)
		}
		so.Key = keyBytes
		certBytes, err := loadFile(certPath)
		if err != nil {
			return SecureOptions{}, errors.WithStack(err)
		}
		so.Certificate = certBytes
	}

	if requireClientCert {
		so.RequireClientCert = true
		var clientRoots [][]byte
		for i := range clientCACertsPath {
			clientCACert, err := loadFile(clientCACertsPath[i])
			if err != nil {
				return SecureOptions{}, err
			}
			clientRoots = append(clientRoots, clientCACert)
		}
		so.ClientRootCAs = clientRoots
	}

	return so, nil
}

func loadFile(path string) ([]byte, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Errorf("Failed opening file %s: %v", path, err)
	}
	return b, nil
}

func newSelfSignedTLSCert(isGm bool) (*tlsgen.CertKeyPair, error) {
	ca, err := tlsgen.NewCA(isGm)
	if err != nil {
		return nil, err
	}
	return ca.NewClientCertKeyPair(isGm)
}
