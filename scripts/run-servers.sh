#!/bin/bash

set -e

MODE=${1:-grpc}
MODE=$(echo "$MODE" | tr '[:upper:]' '[:lower:]')

if [[ "$MODE" != "grpc" && "$MODE" != "rabbitmq" ]]; then
    echo "Usage: $0 [grpc|rabbitmq]"
    exit 1
fi

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$DIR/.." && pwd)"

echo "Starting 3 Authentication Servers in mode: $MODE"
echo "========================================"

cleanup() {
    echo ""
    echo "Stopping all servers..."
    kill $(jobs -p) 2>/dev/null
    exit 0
}

trap cleanup SIGINT SIGTERM

start_server() {
    port=$1
    id=$2
    server_name="${MODE}-server-${id}"
    echo "Starting Server $id on port $port..."
    TRANSPORT=$MODE GRPC_PORT=$port SERVER_ID="$server_name" go run "$ROOT/cmd/server/main.go" &
    pid=$!
    sleep 1
    if ! kill -0 "$pid" 2>/dev/null; then
        echo "Server $id failed to start. See output above."
        cleanup
        exit 1
    fi
    if [[ "$MODE" == "grpc" ]]; then
        if ! wait_for_port "127.0.0.1" "$port"; then
            cleanup
            exit 1
        fi
    fi
    LAST_PID=$pid
}

wait_for_port() {
    host=$1
    port=$2
    for i in {1..20}; do
        if (echo >"/dev/tcp/$host/$port") >/dev/null 2>&1; then
            return 0
        fi
        sleep 0.5
    done
    echo "Server on $host:$port did not become ready. See output above."
    return 1
}

start_server 9001 1
SERVER1_PID=$LAST_PID

#sleep 2

#start_server 9002 2
#SERVER2_PID=$LAST_PID

#sleep 2

#start_server 9003 3
#SERVER3_PID=$LAST_PID

echo ""
echo "All servers started!"
echo "Server 1 PID: $SERVER1_PID (port 9001)"
echo "Server 2 PID: $SERVER2_PID (port 9002)"
echo "Server 3 PID: $SERVER3_PID (port 9003)"
echo "Mode: $MODE"
echo ""
echo "Press Ctrl+C to stop all servers"

wait
