# Build stage
FROM golang:1.24-alpine AS build

# Set working directory
WORKDIR /go/src/streaming-system

# Copy the workspace file first
COPY go.work go.work

# Copy all workspace modules
COPY ClientSideServiceDiscovery/ ClientSideServiceDiscovery/
COPY CoordinateWithConsensus/ CoordinateWithConsensus/
COPY LetsGo/ LetsGo/
COPY SecurityAndObservability/ SecurityAndObservability/
COPY ServeRequestsWithgRPC/ ServeRequestsWithgRPC/
COPY ServerSideServiceDiscovery/ ServerSideServiceDiscovery/
COPY StructureDataWithProtobuf/ StructureDataWithProtobuf/
COPY WriteALogPackage/ WriteALogPackage/
COPY Deploy/ Deploy/

# Change to Deploy directory
WORKDIR /go/src/streaming-system/Deploy

# Set GOPROXY with fallback
ENV GOPROXY=https://proxy.golang.org,direct

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 go build -o /go/bin/streaming-system ./cmd/StreamingSystem

# Download grpc_health_probe
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.34 && \
    wget -qO/go/bin/grpc_health_probe \
    https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/\
${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /go/bin/grpc_health_probe

# Runtime stage
FROM scratch

# Copy the binaries from build stage
COPY --from=build /go/bin/streaming-system /bin/streaming-system
COPY --from=build /go/bin/grpc_health_probe /bin/grpc_health_probe

# Set the entrypoint
ENTRYPOINT ["/bin/streaming-system"]