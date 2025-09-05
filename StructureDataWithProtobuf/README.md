# StructureDataWithProtobuf

This module uses Protocol Buffers to structure data for the Event Streaming System.

## Prerequisites

1. **Go 1.24.0+**
2. **Protocol Buffers Compiler (protoc)**
   ```bash
   # Linux/Ubuntu
   sudo apt install protobuf-compiler
   
   # macOS
   brew install protobuf
   ```

3. **Protobuf Go Plugin**
   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   ```

## Protobuf Compilation

Compile `.proto` files to Go code:

```bash
make compile
```

## Protobuf Go Runtime

Uses `google.golang.org/protobuf` for serialization and deserialization.

## Testing

```bash
make test
```