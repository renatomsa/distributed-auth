#!/bin/bash

# Script para rodar múltiplos servidores gRPC

echo "🚀 Starting 3 gRPC Authentication Servers"
echo "========================================"

# Função para cleanup ao sair
cleanup() {
    echo ""
    echo "🛑 Stopping all servers..."
    kill $(jobs -p) 2>/dev/null
    exit 0
}

trap cleanup SIGINT SIGTERM

# Servidor 1
echo "Starting Server 1 on port 9001..."
GRPC_PORT=9001 SERVER_ID=grpc-server-1 go run cmd/server/main.go &
SERVER1_PID=$!

sleep 2

# Servidor 2
echo "Starting Server 2 on port 9002..."
GRPC_PORT=9002 SERVER_ID=grpc-server-2 go run cmd/server/main.go &
SERVER2_PID=$!

sleep 2

# Servidor 3
echo "Starting Server 3 on port 9003..."
GRPC_PORT=9003 SERVER_ID=grpc-server-3 go run cmd/server/main.go &
SERVER3_PID=$!

echo ""
echo "✅ All servers started!"
echo "Server 1 PID: $SERVER1_PID (port 9001)"
echo "Server 2 PID: $SERVER2_PID (port 9002)"
echo "Server 3 PID: $SERVER3_PID (port 9003)"
echo ""
echo "Press Ctrl+C to stop all servers"

# Esperar indefinidamente
wait
