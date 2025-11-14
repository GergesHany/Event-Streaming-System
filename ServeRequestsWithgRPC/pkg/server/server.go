package server

import (
	"context"
	"time"

	api "github.com/GergesHany/Event-Streaming-System/ServeRequestsWithgRPC/api/v1"
	"google.golang.org/grpc"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"      // for chaining multiple interceptors
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"       // for authentication
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap" // for logging
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"    // for adding context tags to logs

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// ensure grpcServer satisfies the api.LogServer interface
var _ api.LogServer = (*grpcServer)(nil)

type CommitLog interface {
	Append(*api.Record) (uint64, error)
	Read(uint64) (*api.Record, error)
}

type Authorizer interface {
	Authorize(subject, object, action string) error
}

type GetServerer interface {
	GetServers() ([]*api.Server, error)
}

type Config struct {
	CommitLog  CommitLog
	Authorizer Authorizer
	GetServers GetServerer
}

type subjectContextKey struct{}

const (
	objectWildcard = "*"
	produceAction  = "produce"
	consumeAction  = "consume"
)

type grpcServer struct {
	*Config
	// used to ensure forward compatibility when adding methods to the service.
	*api.UnimplementedLogServer
}

func newgrpcServer(config *Config) (srv *grpcServer, err error) {
	srv = &grpcServer{
		Config:                 config,
		UnimplementedLogServer: &api.UnimplementedLogServer{},
	}
	return srv, err
}

func NewGRPCServer(config *Config, opts ...grpc.ServerOption) (*grpc.Server, error) {
	// Set up Zap logger (structured logging)
	logger := zap.L().Named("server")
	zapOpts := []grpc_zap.Option{
		grpc_zap.WithDurationField(func(duration time.Duration) zapcore.Field {
			return zap.Int64("grpc.time_ns", duration.Nanoseconds())
		}),
	}

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	err := view.Register(ocgrpc.DefaultServerViews...)
	if err != nil {
		return nil, err
	}

	// ------------ Initialize the authentication interceptor ------------

	// Add stream interceptor with chained middleware
	opts = append(opts, grpc.StreamInterceptor(
		grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_zap.StreamServerInterceptor(logger, zapOpts...),
			grpc_auth.StreamServerInterceptor(authenticate),
		),
	))

	// Add unary interceptor with chained middleware
	opts = append(opts, grpc.UnaryInterceptor(
		grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(logger, zapOpts...),
			grpc_auth.UnaryServerInterceptor(authenticate),
		),
	))

	// Add stats handler for OpenCensus telemetry
	opts = append(opts, grpc.StatsHandler(&ocgrpc.ServerHandler{}))

	gsrv := grpc.NewServer(opts...)
	srv, err := newgrpcServer(config)

	// Register health check service
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(gsrv, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	if err != nil {
		return nil, err
	}

	api.RegisterLogServer(gsrv, srv)

	return gsrv, nil
}

func (s *grpcServer) Produce(ctx context.Context, req *api.ProduceRequest) (*api.ProduceResponse, error) {
	if err := s.Authorizer.Authorize(subject(ctx), objectWildcard, produceAction); err != nil {
		return nil, err
	}

	offset, err := s.CommitLog.Append(req.Record)
	if err != nil {
		return nil, err
	}
	return &api.ProduceResponse{Offset: offset}, nil
}

func (s *grpcServer) Consume(ctx context.Context, req *api.ConsumeRequest) (*api.ConsumeResponse, error) {
	if err := s.Authorizer.Authorize(subject(ctx), objectWildcard, consumeAction); err != nil {
		return nil, err
	}

	record, err := s.CommitLog.Read(req.Offset)
	if err != nil {
		return nil, api.ErrOffsetOutOfRange{Offset: req.Offset}
	}
	return &api.ConsumeResponse{Record: record}, nil
}

func (s *grpcServer) ProduceStream(stream api.Log_ProduceStreamServer) error {
	for {
		// Recv blocks until it receives a message or an error
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		// Pass the context from the stream to the Produce method
		res, err := s.Produce(stream.Context(), req)
		if err != nil {
			return err
		}
		// Send blocks until the message is sent or an error occurs
		if err := stream.Send(res); err != nil {
			return err
		}
	}
}

func (s *grpcServer) ConsumeStream(req *api.ConsumeRequest, stream api.Log_ConsumeStreamServer) error {
	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
			res, err := s.Consume(stream.Context(), req)
			switch err.(type) {
			case nil:
			case api.ErrOffsetOutOfRange:
				continue
			default:
				return err
			}
			if err = stream.Send(res); err != nil {
				return err
			}
			req.Offset++
		}
	}
}

func (s *grpcServer) GetServers(ctx context.Context, req *api.GetServersRequest) (*api.GetServersResponse, error) {
	servers, err := s.Config.GetServers.GetServers()
	if err != nil {
		return nil, err
	}
	return &api.GetServersResponse{Servers: servers}, nil
}

func authenticate(ctx context.Context) (context.Context, error) {
	// Skip authentication for health check service
	if method, ok := grpc.Method(ctx); ok {
		if method == "/grpc.health.v1.Health/Check" || method == "/grpc.health.v1.Health/Watch" {
			return ctx, nil
		}
	}

	// Extract the peer (client) information from the context
	p, ok := peer.FromContext(ctx)
	if !ok {
		return ctx, status.New(codes.Unknown, "couldn't find p info").Err()
	}

	// If TLS is not configured, allow the connection (development/testing mode)
	// In production, TLS should always be configured
	if p.AuthInfo == nil {
		return ctx, nil
	}

	tlsInfo := p.AuthInfo.(credentials.TLSInfo)                      // Get the TLS information from the AuthInfo
	subject := tlsInfo.State.VerifiedChains[0][0].Subject.CommonName // Extract the Common Name from the verified certificate
	ctx = context.WithValue(ctx, subjectContextKey{}, subject)       // Store the subject in the context

	return ctx, nil
}

func subject(ctx context.Context) string {
	return ctx.Value(subjectContextKey{}).(string)
}
