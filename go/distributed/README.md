# Simple Distributed System Example

This example demonstrates a basic distributed system with a master server and two slave servers.

## Components

- **Master Server**: Coordinates tasks and manages slave servers
- **Slave Servers**: Execute tasks received from the master server

## How It Works

1. The master server starts and waits for slave servers to connect
2. Slave servers connect to the master and register themselves
3. The master assigns tasks to slave servers
4. Slave servers execute tasks and report results back to the master

## Running the Example

Start the master server:
```bash
go run master.go
```

In separate terminals, start two slave servers:
```bash
go run slave.go --id=1 --port=5001
go run slave.go --id=2 --port=5002
```

## Implementation Details

This implementation uses gRPC for communication between servers and demonstrates basic concepts such as:
- Server discovery and registration
- Task distribution
- Health checking
- Simple fault tolerance 