# WriteALogPackage

A simple Write-Ahead Log (WAL) implementation in Go for storing and retrieving records persistently.

## What is it?

A Write-Ahead Log ensures data durability by writing all changes to disk before confirming success. This package provides:

- **Record**: The data you want to store
- **Store**: File that holds the actual data  
- **Index**: File that helps find records quickly
- **Segment**: Combines store + index files
- **Log**: Manages all segments

## Quick Start

```go
package main

import (
    "fmt"
    "os"
    
    logpkg "github.com/GergesHany/Event-Streaming-System/WriteALogPackage/internal/log"
    api "github.com/GergesHany/Event-Streaming-System/StructureDataWithProtobuf/api/v1"
)

func main() {
    // Setup
    dir := "/tmp/mylog"
    os.MkdirAll(dir, 0755)
    defer os.RemoveAll(dir)
    
    config := logpkg.Config{}
    config.Segment.MaxStoreBytes = 1024
    config.Segment.MaxIndexBytes = 1024
    
    // Create log
    myLog, _ := logpkg.NewLog(dir, config)
    defer myLog.Close()
    
    // Write data
    record := &api.Record{Value: []byte("Hello World")}
    offset, _ := myLog.Append(record)
    
    // Read data
    retrieved, _ := myLog.Read(offset)
    fmt.Printf("Retrieved: %s\n", retrieved.Value)
}
```

## Basic Operations

```go
// Create log
myLog, err := logpkg.NewLog("/path/to/log", config)

// Write record
record := &api.Record{Value: []byte("my data")}
offset, err := myLog.Append(record)

// Read record
data, err := myLog.Read(offset)

// Cleanup
myLog.Close()
```

## Configuration

```go
config := logpkg.Config{}
config.Segment.MaxStoreBytes = 1024 * 1024  // 1MB segments
config.Segment.MaxIndexBytes = 1024         // 1KB indexes  
config.Segment.InitialOffset = 0            // Start from 0
```

## Features

- ✅ Fast O(1) read/write operations
- ✅ Thread-safe
- ✅ Automatic file management
- ✅ Data persists across restarts
- ✅ Memory-mapped indexes for speed

## Testing

```bash
go test ./internal/log -v
```