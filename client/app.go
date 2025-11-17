// Package client provides the backend data management and gRPC communication
// layer for the bf-client monitoring tool. It handles connections to Buildfarm
// clusters and maintains application state for operations, workers, and metadata.
package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	reapi "github.com/bazelbuild/remote-apis/build/bazel/remote/execution/v2"
	"google.golang.org/genproto/googleapis/longrunning"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// App represents the main application state and manages gRPC connections to
// the Buildfarm cluster. It tracks operations, workers, metadata, and latency
// metrics across the distributed build system.
type App struct {
	// Instance is the Buildfarm instance name (e.g., "shard").
	Instance string
	// ReapiHost is the Remote Execution API host address.
	ReapiHost string
	// LastRedisLatency tracks the most recent Redis query latency.
	LastRedisLatency time.Duration
	// LastReapiLatency tracks the most recent REAPI call latency.
	LastReapiLatency time.Duration
	// CA is the path to the certificate authority file for TLS connections.
	CA string
	// Done indicates whether the application should terminate.
	Done bool
	// Conn is the main gRPC connection to the REAPI host.
	Conn *grpc.ClientConn
	// workerConns maintains individual gRPC connections to workers.
	workerConns map[string]*grpc.ClientConn
	// Ops stores long-running operations indexed by operation name.
	Ops map[string]*longrunning.Operation
	// Metadatas stores request metadata indexed by operation name.
	Metadatas map[string]*reapi.RequestMetadata
	// Invocations maps invocation IDs to lists of operation names.
	Invocations map[string][]string
	// Fetches counts the number of data fetches performed.
	Fetches uint
	// Mutex protects concurrent access to shared state.
	Mutex *sync.Mutex

	// FrameLimit is the maximum frame rate for UI updates (FPS).
	FrameLimit int
	// SkipFrames is the number of frames to skip before the next update.
	SkipFrames int
	// UpdateCountdown tracks frames until the next data update.
	UpdateCountdown int
}

// NewApp creates and initializes a new App instance with the specified
// REAPI host and optional certificate authority file for TLS.
//
// Parameters:
//   - reapiHost: The Remote Execution API host address (e.g., "localhost:8980" or "grpcs://example.com:443")
//   - ca: Path to the certificate authority file for TLS connections (empty string for default TLS or insecure)
//
// Returns a pointer to the initialized App with default settings including
// a 60 FPS frame limit.
func NewApp(reapiHost string, ca string) *App {
	return &App{
		Instance:    "shard",
		ReapiHost:   reapiHost,
		CA:          ca,
		Done:        false,
		Ops:         make(map[string]*longrunning.Operation),
		Metadatas:   make(map[string]*reapi.RequestMetadata),
		Invocations: make(map[string][]string),
		workerConns: make(map[string]*grpc.ClientConn),
		Mutex:       &sync.Mutex{},
		FrameLimit:  60,
	}
}

// Connect establishes the main gRPC connection to the REAPI host.
// This connection is stored in App.Conn and used for Remote Execution API calls.
// Panics if the connection fails.
func (a *App) Connect() {
	a.Conn = connect(a.ReapiHost, a.CA)
}

// connect establishes a gRPC connection to the specified host with appropriate
// security credentials. It supports both secure (grpcs://) and insecure connections.
//
// For grpcs:// URLs:
//   - Strips the protocol prefix and defaults to port 443 if not specified
//   - Loads TLS credentials from the CA certificate file
//
// For other URLs:
//   - Uses insecure credentials (no TLS)
//
// Parameters:
//   - host: The host address, optionally prefixed with "grpcs://"
//   - ca: Path to the certificate authority file for TLS (empty for default system certs)
//
// Returns the established gRPC connection. Panics on connection failure.
func connect(host string, ca string) *grpc.ClientConn {
	var opts []grpc.DialOption
	if strings.HasPrefix(host, "grpcs://") {
		host = host[8:]
		if !strings.Contains(host, ":") {
			host = host + ":443"
		}
		creds, err := loadTLSCredentials(ca)
		if err != nil {
			panic(err)
		}
		opts = []grpc.DialOption{grpc.WithTransportCredentials(creds)}
	} else {
		opts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	}
	conn, err := grpc.Dial(host, opts...)
	if err != nil {
		panic(err)
	}
	return conn
}

// loadTLSCredentials creates TLS transport credentials for gRPC connections.
// If no CA file is specified, it returns credentials with default TLS configuration.
// Otherwise, it loads the CA certificate from the specified file and creates
// a certificate pool for server certificate verification.
//
// Parameters:
//   - ca: Path to the certificate authority PEM file (empty for default system certs)
//
// Returns:
//   - credentials.TransportCredentials for use with gRPC connections
//   - error if the CA file cannot be read or parsed
func loadTLSCredentials(ca string) (credentials.TransportCredentials, error) {
	if ca == "" {
		return credentials.NewTLS(&tls.Config{}), nil
	}
	// Load certificate of the CA who signed server's certificate
	pemServerCA, err := os.ReadFile(ca)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Create the credentials and return it
	config := &tls.Config{
		RootCAs: certPool,
	}

	return credentials.NewTLS(config), nil
}
