.PHONY: proto run-server run-client test build docker-up docker-down clean help

# Gerar código a partir do protobuf
proto:
	@echo "Generating Go code from protobuf..."
	protoc --go_out=. --go_opt=paths=source_relative \
	    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	    proto/auth.proto
	@echo "Code generated successfully"

# Rodar servidor (porta 9001)
run-server:
	@echo "Starting gRPC server on port 9001..."
	GRPC_PORT=9001 SERVER_ID=grpc-server-1 go run cmd/server/main.go

# Rodar múltiplos servidores
run-servers:
	@echo "Starting 3 gRPC servers..."
	@make run-server1 & make run-server2 & make run-server3

run-server1:
	GRPC_PORT=9001 SERVER_ID=grpc-server-1 go run cmd/server/main.go

run-server2:
	GRPC_PORT=9002 SERVER_ID=grpc-server-2 go run cmd/server/main.go

run-server3:
	GRPC_PORT=9003 SERVER_ID=grpc-server-3 go run cmd/server/main.go

# Rodar cliente de teste
run-client:
	@echo "Running client tests..."
	go run cmd/client/main.go

# Rodar testes unitários
test:
	@echo "Running tests..."
	go test ./... -v -cover

# Build do projeto
build:
	@echo "Building server..."
	go build -o bin/server cmd/server/main.go
	@echo "Building client..."
	go build -o bin/client cmd/client/main.go
	@echo "Build completed"

# Subir docker-compose (PostgreSQL)
docker-up:
	@echo "Starting Docker services..."
	cd deployments && docker-compose up -d
	@echo "Docker services started"
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 5

# Parar docker-compose
docker-down:
	@echo "Stopping Docker services..."
	cd deployments && docker-compose down
	@echo "Docker services stopped"

# Limpar binários e cache
clean:
	@echo "Cleaning..."
	rm -rf bin/
	go clean
	@echo "Clean completed"

# Ajuda
help:
	@echo "Available commands:"
	@echo "  make proto          - Generate Go code from protobuf"
	@echo "  make run-server     - Run single gRPC server (port 9001)"
	@echo "  make run-servers    - Run 3 gRPC servers (ports 9001-9003)"
	@echo "  make run-client     - Run client tests"
	@echo "  make test           - Run unit tests"
	@echo "  make build          - Build server and client binaries"
	@echo "  make docker-up      - Start PostgreSQL in Docker"
	@echo "  make docker-down    - Stop Docker services"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make help           - Show this help message"
