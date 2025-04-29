#!/bin/bash

# Stop any background processes on exit
trap "trap - SIGTERM && kill -- -$$" SIGINT SIGTERM EXIT

# Go to the directory of this script
cd "$(dirname "$0")"

# Install dependencies and generate protobuf code
echo "Setting up dependencies and generating protobuf code..."
make setup

# Start the master server in the background
echo "Starting master server..."
go run master/master.go &
MASTER_PID=$!

# Give the master time to start up
sleep 2

# Start the slave servers in the background
echo "Starting slave 1..."
go run slave/slave.go --id=1 --port=5001 &
SLAVE1_PID=$!

echo "Starting slave 2..."
go run slave/slave.go --id=2 --port=5002 &
SLAVE2_PID=$!

# Wait for a key press
echo ""
echo "Distributed system is running with 1 master and 2 slaves"
echo "Press any key to stop the servers..."
read -n 1

# Clean up
echo "Stopping servers..."
kill $MASTER_PID $SLAVE1_PID $SLAVE2_PID

echo "Done." 