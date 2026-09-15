# Distributed Raft Storage Engine

A high-performance, distributed key-value storage engine built on the Raft consensus algorithm. Supports multi-node clusters, automatic leader election, and persistent storage.

## Features

- **Raft Consensus**: Multi-node coordination with automatic leader election
- **Persistent Storage**: BoltDB-backed log and state snapshots
- **gRPC API**: Efficient serialization with Protocol Buffers
- **Cluster Management**: Add/remove peers dynamically
- **State Snapshots**: Efficient log compaction and recovery
- **Read-Your-Writes Consistency**: Strong consistency guarantees

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Node 1    │     │   Node 2    │     │   Node 3    │
│  (Leader)   │────▶│ (Follower)  │────▶│ (Follower)  │
└─────────────┘     └─────────────┘     └─────────────┘
      │                   │                    │
   Raft Log            Raft Log            Raft Log
   State Machine       State Machine       State Machine
```

### Core Components

- **RaftEngine**: Orchestrates the Raft cluster and state machine
- **StateMachine**: In-memory key-value store with snapshot support
- **gRPC Server**: Handles client requests and cluster communication
- **Log Store**: Persistent Raft log using BoltDB
- **Snapshot Store**: Fast recovery from snapshots

## Quick Start

### Prerequisites

- Go 1.21 or later
- Make (optional)

### Setup

```bash
# Clone the repository
git clone https://github.com/vishu-commits7/distributed-raft-storage.git
cd distributed-raft-storage

# Install dependencies
go mod download

# Generate protobuf code
make protoc
```

### Run a Single Node

```bash
# Terminal 1: Start the first node (bootstrap)
go run cmd/server/main.go -id node1 -addr 127.0.0.1 -port 5000 -grpc-port 6000 -bootstrap

# Terminal 2: Start the second node
go run cmd/server/main.go -id node2 -addr 127.0.0.1 -port 5001 -grpc-port 6001

# Terminal 3: Start the third node
go run cmd/server/main.go -id node3 -addr 127.0.0.1 -port 5002 -grpc-port 6002
```

### Add Peers to Cluster

```bash
go run cmd/client/main.go -addr 127.0.0.1:6000 join -id node2 -addr 127.0.0.1:5001
go run cmd/client/main.go -addr 127.0.0.1:6000 join -id node3 -addr 127.0.0.1:5002
```

### Basic Operations

```bash
# Set a value
go run cmd/client/main.go -addr 127.0.0.1:6000 set key1 value1

# Get a value
go run cmd/client/main.go -addr 127.0.0.1:6000 get key1

# Delete a key
go run cmd/client/main.go -addr 127.0.0.1:6000 delete key1

# Scan keys with prefix
go run cmd/client/main.go -addr 127.0.0.1:6000 scan key
```

## API Reference

### Get

Retrrieves a value from the storage.

```protobuf
rpc Get(GetRequest) returns (GetResponse);

message GetRequest {
  string key = 1;
}

message GetResponse {
  string value = 1;
  bool found = 2;
}
```

### Set

Sets a key-value pair. Only works on the leader.

```protobuf
rpc Set(SetRequest) returns (SetResponse);

message SetRequest {
  string key = 1;
  string value = 2;
}

message SetResponse {
  bool success = 1;
  string error = 2;
}
```

### Delete

Deletes a key. Only works on the leader.

```protobuf
rpc Delete(DeleteRequest) returns (DeleteResponse);

message DeleteRequest {
  string key = 1;
}

message DeleteResponse {
  bool success = 1;
  string error = 2;
}
```

### Scan

Returns all keys with a given prefix (streaming).

```protobuf
rpc Scan(ScanRequest) returns (stream ScanResponse);

message ScanRequest {
  string prefix = 1;
  int32 limit = 2;
}

message ScanResponse {
  string key = 1;
  string value = 2;
}
```

## Configuration

### Command-Line Flags

```
-id string
    Node ID (default "node1")

-addr string
    Bind address (default "127.0.0.1")

-port int
    Raft port (default 5000)

-grpc-port int
    gRPC port (default 6000)

-data string
    Data directory (default "/tmp/raft-node")

-bootstrap
    Bootstrap the cluster (first node only)
```

## Testing

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestStateMachine ./pkg
```

## Performance

- **Throughput**: ~10,000 ops/sec (single node)
- **Latency**: ~5-10ms per write (replicated across 3 nodes)
- **Snapshot Interval**: 8192 log entries (configurable)

## Project Structure

```
.
├── cmd/
│   ├── server/          Server binary
│   └── client/          CLI client binary
├── pkg/
│   ├── raft_engine.go   Core Raft orchestration
│   ├── state_machine.go FSM and snapshots
│   ├── grpc_server.go   gRPC service implementation
│   └── proto/           Generated protobuf code
├── proto/
│   └── storage.proto    Service definitions
├── tests/               Test suite
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Roadmap

- [ ] REST API support
- [ ] Persistence with RocksDB
- [ ] Metrics and monitoring (Prometheus)
- [ ] Admin UI dashboard
- [ ] Backup and restore tools
- [ ] Encryption at rest
- [ ] Multi-region replication

## Contributing

Contributions are welcome! Please open an issue or PR.

## License

MIT License
