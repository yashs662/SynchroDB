# SynchroDB

My attempt at making a distributed KV store

# Proposed architechture

```
SynchroDB/
├── cmd/                        # Main entry points for the application
│   ├── cli/                    # Command-line client for interacting with the database
│   │   └── main.go
│   └── server/                 # Main binary for running the database server
│      └── main.go
├── internal/                   # Private application-specific utilities
│   ├── logger/                 # Logging utilities
│   │   └── logger.go
│   ├── metrics/                # Monitoring and metrics collection
│   │   └── metrics.go
│   └── utils/                  # Miscellaneous utilities
│       └── helpers.go
├── pkg/                        # Library code shared across the project
│   ├── cluster/                # Distributed system utilities
│   │   ├── consensus/          # Consensus protocol (Raft or custom)
│   │   │   ├── leader_election.go
│   │   │   └── replication.go
│   │   ├── node.go             # Node configuration and management
│   │   └── sharding.go         # Sharding logic
│   ├── database/               # Core database logic
│   │   ├── kvstore.go          # In-memory key-value store
│   │   ├── replication.go      # Replication logic for distributed systems
│   │   └── persistence.go      # Optional: File-based persistence
│   ├── protocol/               # Custom Redis-style protocol handling
│   │   ├── server.go           # Protocol server implementation
│   │   └── parser.go           # Protocol parsing logic
│   └── grpc/                   # gRPC API implementation (future distributed features)
│       ├── service.go          # gRPC service definitions and implementation
│       └── protobuf/           # Auto-generated Protobuf files
│           ├── database.pb.go
│           └── database_grpc.pb.go
├── tests/                      # Integration and end-to-end tests (future)
│   ├── single_node_test.go     # Tests for single-node operations
│   ├── cluster_test.go         # Tests for distributed setup
│   └── protocol_test.go        # Tests for the custom protocol
├── Dockerfile                  # Dockerfile for containerizing the database
├── Makefile                    # Makefile for build, test, and run tasks
└── README.md                   # Documentation for the project
```
