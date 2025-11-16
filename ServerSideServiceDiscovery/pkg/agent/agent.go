package agent

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"

	DisLog "github.com/GergesHany/Event-Streaming-System/CoordinateWithConsensus/pkg/log"
	"github.com/GergesHany/Event-Streaming-System/SecurityAndObservability/pkg/auth"
	api "github.com/GergesHany/Event-Streaming-System/ServeRequestsWithgRPC/api/v1"
	"github.com/GergesHany/Event-Streaming-System/ServeRequestsWithgRPC/pkg/server"
	"github.com/GergesHany/Event-Streaming-System/ServerSideServiceDiscovery/pkg/discovery"
	"github.com/GergesHany/Event-Streaming-System/WriteALogPackage/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/soheilhy/cmux"

	"bytes"
	"io"

	"github.com/hashicorp/raft"
)

type Agent struct {
	Config

	mux cmux.CMux // Connection multiplexer

	log        *DisLog.DistributedLog // Distributed log using Raft consensus
	server     *grpc.Server
	membership *discovery.Membership // Service discovery membership

	// Shutdown coordination
	shutdown     bool
	shutdowns    chan struct{}
	shutdownLock sync.Mutex
}

type Config struct {
	ServerTLSConfig *tls.Config
	PeerTLSConfig   *tls.Config

	DataDir  string
	BindAddr string

	RPCPort int

	NodeName       string
	StartJoinAddrs []string

	ACLModelFile  string
	ACLPolicyFile string

	Bootstrap bool
}

func (c Config) RPCAddr() (string, error) {
	host, _, err := net.SplitHostPort(c.BindAddr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", host, c.RPCPort), nil
}

func New(config Config) (*Agent, error) {
	a := &Agent{
		Config:    config,
		shutdowns: make(chan struct{}),
	}

	setup := []func() error{
		a.setupMux,
		a.setupLog,
		a.setupServer,
		a.setupMembership,
	}

	// Execute all setup functions
	for _, fn := range setup {
		if err := fn(); err != nil {
			return nil, err
		}
	}

	go a.serve()
	return a, nil
}

// func (a *Agent) setupLogger() error {
// 	logger, err := zap.NewDevelopment()
// 	if err != nil {
// 		return err
// 	}
// 	// Replace the global logger with the new logger
// 	zap.ReplaceGlobals(logger)
// 	return nil
// }

func (a *Agent) setupLog() error {
	raftLn := a.mux.Match(func(reader io.Reader) bool {
		b := make([]byte, 1)
		if _, err := reader.Read(b); err != nil {
			return false
		}
		return bytes.Equal(b, []byte{byte(log.RaftRPC)})
	})

	logConfig := log.Config{}
	logConfig.Segment.MaxStoreBytes = 1024 * 1024 * 1024 // 1GB
	logConfig.Segment.MaxIndexBytes = 1024 * 1024        // 1MB
	logConfig.Raft.StreamLayer = log.NewStreamLayer(
		raftLn,
		a.Config.ServerTLSConfig,
		a.Config.PeerTLSConfig,
	)

	rpcAddr, err := a.Config.RPCAddr()
	if err != nil {
		return err
	}

	logConfig.Raft.BindAddr = rpcAddr
	logConfig.Raft.LocalID = raft.ServerID(a.Config.NodeName)
	logConfig.Raft.Bootstrap = a.Config.Bootstrap

	a.log, err = DisLog.NewDistributedLog(
		a.Config.DataDir,
		logConfig,
	)

	if err != nil {
		return err
	}

	// Don't wait for leader during setup - it will be elected asynchronously
	// after all nodes have joined via Serf membership
	return err
}

func (a *Agent) setupServer() error {
	authorizer := auth.New(a.Config.ACLModelFile, a.Config.ACLPolicyFile)
	serverConfig := &server.Config{
		CommitLog:  a.log,
		Authorizer: authorizer,
		GetServers: a, // Agent implements GetServers interface
	}

	var opts []grpc.ServerOption
	if a.Config.ServerTLSConfig != nil {
		creds := credentials.NewTLS(a.Config.ServerTLSConfig)
		opts = append(opts, grpc.Creds(creds))
	}

	var err error
	a.server, err = server.NewGRPCServer(serverConfig, opts...)
	if err != nil {
		return err
	}

	grpcLn := a.mux.Match(cmux.Any()) // Match all other connections for gRPC

	// Start the gRPC server in a separate goroutine
	go func() {
		if err := a.server.Serve(grpcLn); err != nil {
			_ = a.Shutdown() // Shutdown on server error
		}
	}()

	return err
}

func (a *Agent) setupMembership() error {
	rpcAddr, err := a.Config.RPCAddr()
	if err != nil {
		return err
	}

	a.membership, err = discovery.New(a.log, discovery.Config{
		NodeName:       a.Config.NodeName,
		BindAddr:       a.Config.BindAddr,
		Tags:           map[string]string{"rpc_addr": rpcAddr},
		StartJoinAddrs: a.Config.StartJoinAddrs,
	})

	return err
}

func (a *Agent) Shutdown() error {

	a.shutdownLock.Lock()
	defer a.shutdownLock.Unlock()

	if a.shutdown {
		return nil
	}

	a.shutdown = true
	close(a.shutdowns)

	// Define shutdown functions to be called in order
	shutdowns := []func() error{
		a.membership.Leave,
		func() error {
			// Gracefully stop the gRPC server
			a.server.GracefulStop()
			return nil
		},
		a.log.Close,
	}

	for _, fn := range shutdowns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func (a *Agent) setupMux() error {
	// Bind RPC server to all interfaces (0.0.0.0) to accept external connections
	// Parse BindAddr to extract just the port if needed, but bind mux to 0.0.0.0
	rpcAddr := fmt.Sprintf("0.0.0.0:%d", a.Config.RPCPort)
	ln, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		return err
	}

	a.mux = cmux.New(ln)
	return nil
}

// GetServers implements the server.GetServerer interface by adapting
// the distributed log's GetServers to the gRPC API's Server type
func (a *Agent) GetServers() ([]*api.Server, error) {
	servers, err := a.log.GetServers()
	if err != nil {
		return nil, err
	}
	// Convert from distributed log Server type to API Server type
	apiServers := make([]*api.Server, len(servers))
	for i, srv := range servers {
		apiServers[i] = &api.Server{
			Id:       srv.ID,
			RpcAddr:  srv.Address,
			IsLeader: srv.IsLeader,
		}
	}
	return apiServers, nil
}

func (a *Agent) serve() error {
	if err := a.mux.Serve(); err != nil {
		_ = a.Shutdown()
		return err
	}
	return nil
}
