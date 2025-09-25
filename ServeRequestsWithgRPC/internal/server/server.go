package server

import (
	"context"

	api "github.com/GergesHany/Event-Streaming-System/ServeRequestsWithgRPC/api/v1"
	"google.golang.org/grpc"


	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
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

type Config struct {
	CommitLog  CommitLog
	Authorizer Authorizer
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
		Config: config,
		UnimplementedLogServer: &api.UnimplementedLogServer{},
	}
	return srv, err
}

func NewGRPCServer(config *Config, opts ...grpc.ServerOption) (*grpc.Server, error) {
	// Add the authentication interceptor (is basically a middleware) to the gRPC server options
	Chain := grpc_middleware.ChainStreamServer(grpc_auth.StreamServerInterceptor(authenticate))

	// Add the unary interceptor (is basically a middleware) to the gRPC server options
	ChainUnary := grpc_middleware.ChainUnaryServer(grpc_auth.UnaryServerInterceptor(authenticate))

	// Add the stream interceptor (is basically a middleware) to the gRPC server options
	opts = append(opts, grpc.StreamInterceptor(Chain), grpc.UnaryInterceptor(ChainUnary))

	gsrv := grpc.NewServer(opts...)
	srv, err := newgrpcServer(config)
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

func authenticate(ctx context.Context) (context.Context, error) {
	// Extract the peer (client) information from the context
	p, ok := peer.FromContext(ctx)
	if !ok {
		return ctx, status.New(codes.Unknown, "couldn't find p info").Err()
	}

	// Ensure that transport security is being used
	if p.AuthInfo == nil {
		return ctx, status.New(codes.Unauthenticated, "no transport security being used").Err()
	}

	tlsInfo := p.AuthInfo.(credentials.TLSInfo) // Get the TLS information from the AuthInfo
	subject := tlsInfo.State.VerifiedChains[0][0].Subject.CommonName // Extract the Common Name from the verified certificate
	ctx = context.WithValue(ctx, subjectContextKey{}, subject) // Store the subject in the context

	return ctx, nil
}

func subject(ctx context.Context) (string) {
	return ctx.Value(subjectContextKey{}).(string)
}