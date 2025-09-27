package server

import (
	"context"
	"flag"
	"net"
	"os"
	"time"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc/credentials"

	api "github.com/GergesHany/Event-Streaming-System/ServeRequestsWithgRPC/api/v1"
	"github.com/GergesHany/Event-Streaming-System/WriteALogPackage/log"

	auth "github.com/GergesHany/Event-Streaming-System/SecurityAndObservability/pkg/auth"
	SecureConfig "github.com/GergesHany/Event-Streaming-System/SecurityAndObservability/pkg/config"

	"go.opencensus.io/examples/exporter"
	"go.uber.org/zap"
)

var debug = flag.Bool("debug", false, "Enable observability for debugging.")

func TestMain(m *testing.M) {
	flag.Parse()
	if *debug {
		logger, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		zap.ReplaceGlobals(logger)
	}
	os.Exit(m.Run())
}

func TestServer(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T, rootClient api.LogClient, nobodyClient api.LogClient, config *Config){
		"produce/consume a message to/from the log succeeds": testProduceConsume,
		"produce/consume stream succeeds":                    testProduceConsumeStream,
		"consume past log boundary fails":                    testConsumePastBoundary,
		"unauthorized fails":                                 testUnauthorized,
	} {
		t.Run(scenario, func(t *testing.T) {
			rootClient, nobodyClient, cfg, teardown := setupTest(t, nil)
			defer teardown()
			fn(t, rootClient, nobodyClient, cfg)
		})
	}
}

func setupTest(t *testing.T, fn func(*Config)) (rootClient api.LogClient, nobodyClient api.LogClient, cfg *Config, teardown func()) {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	// Helper function to create a client with given certificate and key files
	newClient := func(crtPath, keyPath string) (*grpc.ClientConn, api.LogClient, []grpc.DialOption) {
		tlsConfig, err := SecureConfig.SetupTLSConfig(SecureConfig.TLSConfig{
			CertFile: crtPath,
			KeyFile:  keyPath,
			CAFile:   SecureConfig.CAFile,
			Server:   false,
		})
		require.NoError(t, err)
		tlsCreds := credentials.NewTLS(tlsConfig)
		opts := []grpc.DialOption{grpc.WithTransportCredentials(tlsCreds)}
		conn, err := grpc.Dial(l.Addr().String(), opts...)
		require.NoError(t, err)
		client := api.NewLogClient(conn)
		return conn, client, opts
	}

	// Create clients for "root" and "nobody" users
	var rootConn *grpc.ClientConn
	rootConn, rootClient, _ = newClient(SecureConfig.RootClientCertFile, SecureConfig.RootClientKeyFile)

	var nobodyConn *grpc.ClientConn
	nobodyConn, nobodyClient, _ = newClient(SecureConfig.NobodyClientCertFile, SecureConfig.NobodyClientKeyFile)

	// Set up the server TLS configuration
	serverTLSConfig, err := SecureConfig.SetupTLSConfig(SecureConfig.TLSConfig{
		CertFile:      SecureConfig.ServerCertFile,
		KeyFile:       SecureConfig.ServerKeyFile,
		CAFile:        SecureConfig.CAFile,
		ServerAddress: l.Addr().String(),
		Server:        true,
	})

	require.NoError(t, err)
	serverCreds := credentials.NewTLS(serverTLSConfig)

	dir, err := os.MkdirTemp("", "server-test")
	require.NoError(t, err)

	clog, err := log.NewLog(dir, log.Config{})
	require.NoError(t, err)
	a := auth.New(SecureConfig.ACLModelFile, SecureConfig.ACLPolicyFile)

	// Set up telemetry exporter to log traces and metrics to files for debugging
	var telemetryExporter *exporter.LogExporter
	if *debug {
		metricsLogFile, err := os.CreateTemp("", "metrics-*.log")
		require.NoError(t, err)
		t.Logf("metrics log file: %s", metricsLogFile.Name())

		tracesLogFile, err := os.CreateTemp("", "traces-*.log")
		require.NoError(t, err)
		t.Logf("traces log file: %s", tracesLogFile.Name())

		telemetryExporter, err = exporter.NewLogExporter(exporter.Options{
			MetricsLogFile:    metricsLogFile.Name(),
			TracesLogFile:     tracesLogFile.Name(),
			ReportingInterval: time.Second,
		})
		require.NoError(t, err)
		err = telemetryExporter.Start()
		require.NoError(t, err)
	}

	// Create an adapter to convert between Record types
	adapter := NewLogAdapter(clog)

	cfg = &Config{
		CommitLog:  adapter,
		Authorizer: a,
	}

	if fn != nil {
		fn(cfg)
	}
	server, err := NewGRPCServer(cfg, grpc.Creds(serverCreds))
	require.NoError(t, err)

	go func() {
		server.Serve(l)
	}()

	return rootClient, nobodyClient, cfg, func() {
		server.Stop()
		rootConn.Close()
		nobodyConn.Close()
		l.Close()

		/*
		  * We sleep for 1.5 seconds to give the telemetry exporter enough time to flush its data to disk. 
		  * Then we stop and close the exporter.
		*/
		if telemetryExporter != nil {
			time.Sleep(1500 * time.Millisecond)
			telemetryExporter.Stop()
			telemetryExporter.Close()
		}
	}
}

func testProduceConsume(t *testing.T, client, _ api.LogClient, config *Config) {
	ctx := context.Background()

	want := &api.Record{Value: []byte("hello world")}

	produce, err := client.Produce(
		ctx,
		&api.ProduceRequest{Record: want},
	)

	require.NoError(t, err)

	consume, err := client.Consume(ctx, &api.ConsumeRequest{
		Offset: produce.Offset,
	})

	require.NoError(t, err)
	require.Equal(t, want.Value, consume.Record.Value)
	require.Equal(t, want.Offset, consume.Record.Offset)
}

func testConsumePastBoundary(t *testing.T, client, _ api.LogClient, config *Config) {
	ctx := context.Background()

	produce, err := client.Produce(ctx, &api.ProduceRequest{
		Record: &api.Record{Value: []byte("hello world")},
	})

	require.NoError(t, err)

	consume, err := client.Consume(ctx, &api.ConsumeRequest{
		Offset: produce.Offset + 1,
	})

	if consume != nil {
		t.Fatal("consume not nil")
	}

	got := status.Code(err)
	want := codes.NotFound
	if got != want {
		t.Fatalf("got err: %v, want: %v", got, want)
	}
}

func testProduceConsumeStream(t *testing.T, client, _ api.LogClient, config *Config) {
	ctx := context.Background()

	// Prepare a list of records to produce
	records := []*api.Record{{
		Value:  []byte("first message"),
		Offset: 0,
	}, {
		Value:  []byte("second message"),
		Offset: 1,
	}, {
		Value:  []byte("third message"),
		Offset: 2,
	}, {
		Value:  []byte("fourth message"),
		Offset: 3,
	}, {
		Value:  []byte("fifth message"),
		Offset: 4,
	}}

	//  ------ Produce Streaming  ------
	stream, err := client.ProduceStream(ctx)
	require.NoError(t, err)

	for offset, record := range records {
		err := stream.Send(&api.ProduceRequest{Record: record})
		require.NoError(t, err)
		res, err := stream.Recv()
		require.NoError(t, err)
		if res.Offset != uint64(offset) {
			t.Fatalf("got offset: %d, want: %d", res.Offset, offset)
		}
	}

	// ------ Consume Streaming ------
	consumeStream, err := client.ConsumeStream(ctx, &api.ConsumeRequest{Offset: 0})
	require.NoError(t, err)

	for i, record := range records {
		res, err := consumeStream.Recv()
		require.NoError(t, err)
		if res.Record.Offset != uint64(i) {
			t.Fatalf("got offset: %d, want: %d", res.Record.Offset, i)
		}
		if string(res.Record.Value) != string(record.Value) {
			t.Fatalf("got value: %s, want: %s", res.Record.Value, record.Value)
		}
	}
}

func testUnauthorized(t *testing.T, _, client api.LogClient, config *Config) {
	ctx := context.Background()
	produce, err := client.Produce(ctx,
		&api.ProduceRequest{Record: &api.Record{Value: []byte("hello world")}},
	)

	if produce != nil {
		t.Fatalf("produce response should be nil")
	}

	gotCode, wantCode := status.Code(err), codes.PermissionDenied

	if gotCode != wantCode {
		t.Fatalf("got code: %d, want: %d", gotCode, wantCode)
	}

	consume, err := client.Consume(ctx, &api.ConsumeRequest{
		Offset: 0,
	})

	if consume != nil {
		t.Fatalf("consume response should be nil")
	}

	gotCode, wantCode = status.Code(err), codes.PermissionDenied

	if gotCode != wantCode {
		t.Fatalf("got code: %d, want: %d", gotCode, wantCode)
	}
}
