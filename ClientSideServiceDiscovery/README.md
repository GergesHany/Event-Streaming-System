# Client-Side Service Discovery

This module implements client-side service discovery and load balancing for the Event Streaming System using gRPC's resolver and balancer interfaces.

## Overview

Client-side service discovery allows gRPC clients to automatically discover available server instances and intelligently route requests based on server roles (leader/follower). This approach provides:

- **Automatic server discovery**: Clients query a discovery service to get available servers
- **Smart load balancing**: Routes produce requests to leaders and consume requests to followers
- **Dynamic updates**: Clients automatically adapt to topology changes
- **No proxy overhead**: Direct client-to-server communication

## Architecture

### Components

1. **Resolver (`resolver.go`)**: Discovers available servers and their roles
   - Queries the GetServers API to retrieve server addresses
   - Identifies leader and follower nodes
   - Updates client connection state when topology changes

2. **Picker (`picker.go`)**: Implements load balancing logic
   - Routes `Produce` requests to the leader server
   - Round-robins `Consume` requests across follower servers
   - Handles no available connection scenarios

## How It Works

### Service Discovery Flow

```
┌─────────────┐         ┌──────────────┐         ┌────────────┐
│   Client    │ ──────> │   Resolver   │ ──────> │  Discovery │
│             │  Build  │              │ Query   │  Service   │
└─────────────┘         └──────────────┘         └────────────┘
       │                       │                         │
       │                       │ <───────────────────────┘
       │                       │   Server List
       │                       │   (leader + followers)
       │                       │
       │ <─────────────────────┘
       │    Updated Addresses
       │
       ▼
┌─────────────┐
│   Picker    │  ──> Route Produce to Leader
│             │  ──> Round-robin Consume to Followers
└─────────────┘
```

### Load Balancing Strategy

- **Produce Operations**: Always routed to the leader node to maintain consistency
- **Consume Operations**: Round-robin distributed across follower nodes for load distribution
- **Fallback**: If no followers available, consume requests go to the leader

## Usage

### Client Setup

```go
import (
    "google.golang.org/grpc"
    "github.com/GergesHany/Event-Streaming-System/ClientSideServiceDiscovery/pkg/loadbalance"
)

// Create a client connection with custom resolver
conn, err := grpc.Dial(
    fmt.Sprintf(
        "%s:///%s",
        loadbalance.Name,           // "StreamingSystem"
        "discovery-service:8080",    // Discovery service address
    ),
    grpc.WithDefaultServiceConfig(
        fmt.Sprintf(`{"loadBalancingConfig":[{"%s":{}}]}`, loadbalance.Name),
    ),
)
```

### Registration

Both the resolver and picker are automatically registered via `init()` functions:

```go
func init() {
    resolver.Register(&Resolver{})
}

func init() {
    balancer.Register(
        base.NewBalancerBuilder(Name, &Picker{}, base.Config{}),
    )
}
```

## Implementation Details

### Resolver

The `Resolver` implements the `resolver.Resolver` interface:

- **Build**: Establishes connection to discovery service
- **ResolveNow**: Queries `GetServers` API and updates client state
- **Close**: Cleans up resolver connections

Server attributes include:
- `Addr`: RPC address (e.g., "localhost:9001")
- `is_leader`: Boolean indicating leader status

### Picker

The `Picker` implements `base.PickerBuilder` and `balancer.Picker`:

- **Build**: Separates leader and follower sub-connections
- **Pick**: Routes requests based on method name
- **nextFollower**: Round-robin selection using atomic counter