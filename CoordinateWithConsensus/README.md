# CoordinateWithConsensus

This module implements a distributed, replicated log system using the Raft consensus algorithm. It provides strong consistency guarantees across multiple nodes in a distributed system.

## Overview

The `CoordinateWithConsensus` package builds upon the previous log implementation by adding distributed coordination and consensus capabilities. It uses HashiCorp's Raft library to ensure all nodes in a cluster agree on the log's state.

## Features

- **Distributed Log Replication**: Automatically replicates log entries across multiple nodes
- **Consensus Algorithm**: Uses Raft to ensure all nodes agree on the log state
- **Leader Election**: Automatic leader election when the current leader fails
- **Fault Tolerance**: Continues operation as long as a majority of nodes are available
- **Service Discovery Integration**: Works with the membership management system

## Components

### Distributed Log (`pkg/log/distributed.go`)

The core distributed log implementation that integrates with Raft:

- **DistributedLog**: Main structure that wraps the log with Raft consensus
- **FSM (Finite State Machine)**: Applies commands to the log
- **Log Store**: Persists Raft commands
- **Stable Store**: Stores cluster configuration
- **Snapshot Store**: Creates and restores compact snapshots
- **Transport Layer**: Handles communication between Raft peers

### Membership (`pkg/discovery/membership.go`)

Provides membership management utilities with Raft-aware error handling for service discovery integration.

## Architecture

The Raft instance comprises five key components:

1. **Finite-State Machine**: Applies commands to the log
2. **Log Store**: Stores Raft commands
3. **Stable Store**: Stores cluster configuration (servers, addresses)
4. **Snapshot Store**: Stores compact snapshots for efficient recovery
5. **Transport**: Connects with peer servers in the cluster

## Usage

```go
import (
    "github.com/GergesHany/Event-Streaming-System/CoordinateWithConsensus/pkg/log"
)

// Create a new distributed log
config := log.Config{}
config.Raft.LocalID = "node-1"
config.Raft.Bootstrap = true // Only for the first node

distributedLog, err := log.NewDistributedLog("/data/dir", config)
if err != nil {
    // Handle error
}

// Append to the distributed log
record := &api.Record{Value: []byte("data")}
offset, err := distributedLog.Append(record)

// Read from the distributed log
record, err = distributedLog.Read(offset)
```

## Testing

The module includes comprehensive tests for multi-node scenarios:

```bash
go test ./pkg/log/...
```

Tests verify:
- Multi-node cluster formation
- Leader election
- Log replication across nodes
- Fault tolerance and recovery

## How It Works

1. **Bootstrap**: The first node bootstraps the cluster
2. **Join**: Additional nodes join the existing cluster
3. **Replication**: Leader replicates log entries to followers
4. **Consensus**: Entries are committed once replicated to a majority
5. **Leader Election**: If leader fails, followers elect a new leader

## Configuration

Key configuration options:

- `LocalID`: Unique identifier for the node
- `Bootstrap`: Whether this node should bootstrap a new cluster
- `HeartbeatTimeout`: Timeout for heartbeat messages
- `ElectionTimeout`: Timeout before starting an election
- `StreamLayer`: Network transport layer for Raft
