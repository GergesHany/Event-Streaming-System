# Server-Side Service Discovery

This module implements server-side service discovery for the Event Streaming System using HashiCorp Serf for cluster membership management and automatic log replication across distributed nodes.

## Overview

The Server-Side Service Discovery module provides:

- **Cluster Membership Management**: Automatic discovery and monitoring of cluster nodes using HashiCorp Serf
- **Log Replication**: Automatic replication of logs across all cluster members
- **Agent Management**: A unified agent that coordinates gRPC servers, membership, and replication
- **Fault Tolerance**: Handles node joins, leaves, and failures gracefully

## Components

### 1. Membership (`pkg/discovery/membership.go`)
- Manages cluster membership using HashiCorp Serf
- Handles node joins and leaves
- Provides event-driven notifications for membership changes
- Supports node tagging for metadata (e.g., service endpoints)

### 2. Agent (`pkg/agent/agent.go`)
- Main coordinator that ties together all components
- Manages gRPC server lifecycle
- Coordinates membership and replication
- Handles TLS configuration for secure communication
- Supports graceful shutdown

### 3. Replicator (`pkg/log/replicator.go`)
- Automatically replicates logs to newly joined cluster members
- Maintains connections to all active nodes
- Handles connection failures and retries
- Ensures data consistency across the cluster

## Key Features

- **Automatic Service Discovery**: Nodes automatically discover each other when joining the cluster
- **Dynamic Scaling**: New nodes can join the cluster seamlessly
- **Data Consistency**: Logs are automatically replicated to maintain consistency
- **Security**: Support for TLS encryption and ACL-based authorization
- **Monitoring**: Built-in logging and observability

## Configuration

The agent accepts the following configuration:

```go
type Config struct {
    ServerTLSConfig *tls.Config   // TLS config for server connections
    PeerTLSConfig   *tls.Config   // TLS config for peer connections
    DataDir         string        // Directory for data storage
    BindAddr        string        // Address to bind the service
    RPCPort         int          // Port for RPC communication
    NodeName        string        // Unique name for this node
    StartJoinAddrs  []string     // Addresses of existing nodes to join
    ACLModelFile    string       // ACL model file path
    ACLPolicyFile   string       // ACL policy file path
}
```

## Dependencies

- **HashiCorp Serf**: For cluster membership and failure detection
- **gRPC**: For inter-service communication
- **Zap**: For structured logging
- **Event Streaming System components**: Integrates with other modules in the system

## Usage

This module is designed to be used as part of the larger Event Streaming System. It provides the foundation for building a distributed, fault-tolerant log streaming service where multiple nodes can join a cluster and automatically participate in log replication.

## Testing

The module includes comprehensive tests for:
- Membership management scenarios
- Agent lifecycle management
- Network partition handling
- Replication consistency

Run tests with:
```bash
go test ./...
```