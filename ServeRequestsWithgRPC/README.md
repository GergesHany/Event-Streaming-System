# ServeRequestsWithgRPC

A gRPC-based log service implementation that provides efficient log data streaming capabilities using Protocol Buffers.

## Overview

This project implements a distributed log service using gRPC and Protocol Buffers. It provides both unary and streaming RPC methods for producing and consuming log records with high performance and type safety.

## Features

- **Unary RPCs**: Simple request-response operations for single record operations
  - `Produce`: Add a single record to the log
  - `Consume`: Read a single record from the log

- **Streaming RPCs**: Efficient bulk operations
  - `ConsumeStream`: Server-side streaming for reading multiple records
  - `ProduceStream`: Bidirectional streaming for real-time log operations

- **Protocol Buffer Schema**: Strongly typed data structures
- **Offset-based Access**: Efficient record retrieval using numeric offsets


## Protocol Buffer Schema

The service defines the following core types:

- **Record**: Contains log data with `value` (bytes) and `offset` (uint64)
- **ProduceRequest/Response**: For adding records to the log
- **ConsumeRequest/Response**: For reading records from the log

## Prerequisites

- Go 1.24.0 or later
- Protocol Buffers compiler (`protoc`)
- gRPC Go plugin


## Usage

### Building

Generate Protocol Buffer code from the `.proto` files:

```bash
make compile
```

### Testing

Run the test suite:

```bash
go test ./...
```

## API Reference

### Service Methods

- `Produce(ProduceRequest) returns (ProduceResponse)` - Add a record to the log
- `Consume(ConsumeRequest) returns (ConsumeResponse)` - Read a record from the log
- `ConsumeStream(ConsumeRequest) returns (stream ConsumeResponse)` - Stream multiple records
- `ProduceStream(stream ProduceRequest) returns (stream ProduceResponse)` - Bidirectional streaming


## Dependencies

- [gRPC-Go](https://google.golang.org/grpc): High-performance RPC framework
- [Protocol Buffers](https://google.golang.org/protobuf): Serialization library
- Custom log package: Underlying log storage implementation
