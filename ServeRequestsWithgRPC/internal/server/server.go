package server

import (
	"context"

	api "github.com/GergesHany/Event-Streaming-System/ServeRequestsWithgRPC/api/v1"
	"google.golang.org/grpc"
)

// ensure grpcServer satisfies the api.LogServer interface
var _ api.LogServer = (*grpcServer)(nil)

type CommitLog interface {
	Append(*api.Record) (uint64, error)
	Read(uint64) (*api.Record, error)
}

type Config struct {
	CommitLog CommitLog
}

type grpcServer struct {
	// used to ensure forward compatibility when adding methods to the service.
	api.UnimplementedLogServer
	*Config
}

func newgrpcServer(config *Config) (srv *grpcServer, err error) {
	srv = &grpcServer{
		Config: config,
	}
	return srv, err
}

func (s *grpcServer) Produce(ctx context.Context, req *api.ProduceRequest) (*api.ProduceResponse, error) {
	offset, err := s.CommitLog.Append(req.Record)
	if err != nil {
		return nil, err
	}
	return &api.ProduceResponse{Offset: offset}, nil
}

func (s *grpcServer) Consume(ctx context.Context, req *api.ConsumeRequest) (*api.ConsumeResponse, error) {
	record, err := s.CommitLog.Read(req.Offset)
	if err != nil {
		return nil, err
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

func NewGRPCServer(config *Config) (*grpc.Server, error) {
	gsrv := grpc.NewServer()
	srv, err := newgrpcServer(config)

	if err != nil {
		return nil, err
	}

	api.RegisterLogServer(gsrv, srv)
	return gsrv, nil
}
