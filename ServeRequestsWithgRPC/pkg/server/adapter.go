package server

import (
	"strings"

	grpcapi "github.com/GergesHany/Event-Streaming-System/ServeRequestsWithgRPC/api/v1"
	logapi "github.com/GergesHany/Event-Streaming-System/StructureDataWithProtobuf/api/v1"
	"github.com/GergesHany/Event-Streaming-System/WriteALogPackage/log"
)

// LogAdapter adapts the WriteALogPackage log to work with the gRPC server's CommitLog interface
type LogAdapter struct {
	log *log.Log
}

// NewLogAdapter creates a new LogAdapter
func NewLogAdapter(l *log.Log) *LogAdapter {
	return &LogAdapter{log: l}
}

// Append converts a gRPC Record to a log Record and appends it
func (a *LogAdapter) Append(record *grpcapi.Record) (uint64, error) {
	// Convert gRPC Record to log Record
	logRecord := &logapi.Record{
		Value:  record.Value,
		Offset: record.Offset,
	}
	return a.log.Append(logRecord)
}

// Read reads a record and converts it from log Record to gRPC Record
func (a *LogAdapter) Read(offset uint64) (*grpcapi.Record, error) {
	// Read from the log
	logRecord, err := a.log.Read(offset)
	if err != nil {
		// Convert log package errors to gRPC errors
		if strings.Contains(err.Error(), "offset out of range") {
			return nil, grpcapi.ErrOffsetOutOfRange{Offset: offset}.GRPCStatus().Err()
		}
		return nil, err
	}

	// Convert log Record to gRPC Record
	grpcRecord := &grpcapi.Record{
		Value:  logRecord.Value,
		Offset: logRecord.Offset,
	}
	return grpcRecord, nil
}
