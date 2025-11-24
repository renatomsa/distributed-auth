package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"google.golang.org/grpc"

	"github.com/renatomsa/auth-grpc/internal/auth"
	"github.com/renatomsa/auth-grpc/internal/database"
	grpcserver "github.com/renatomsa/auth-grpc/internal/grpc"
	"github.com/renatomsa/auth-grpc/internal/mq"
	"github.com/renatomsa/auth-grpc/pkg/config"
)

func main() {
	cfg := config.Load()

	log.Println("Starting Auth Server...")
	cfg.Print()

	db, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to PostgreSQL")

	if err := database.RunMigrations(db.GetDB()); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	if cfg.Environment == "development" {
		if err := database.SeedUsers(db.GetDB()); err != nil {
			log.Printf("Warning: Failed to seed database: %v", err)
		}
	}

	authService := auth.NewService(db)
	log.Println("Auth service initialized")

	transport := strings.ToLower(os.Getenv("TRANSPORT"))
	if transport == "" {
		transport = "grpc"
	}

	switch transport {
	case "rabbitmq":
		startRabbitMQ(authService, cfg)
	default:
		startGRPC(authService, cfg)
	}
}

func startGRPC(authService *auth.Service, cfg *config.Config) {
	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024),
		grpc.MaxSendMsgSize(10*1024*1024),
	)

	grpcserver.RegisterServer(grpcServer, authService, cfg.ServerID)
	log.Printf("gRPC server registered on port %s", cfg.Port)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh

		log.Println("\nShutting down gRPC server gracefully...")
		grpcServer.GracefulStop()
		log.Println("gRPC server stopped")
	}()

	log.Printf("gRPC Server [%s] listening on port %s", cfg.ServerID, cfg.Port)
	log.Println("Test credentials: username=alice, password=password123")
	log.Println("Press Ctrl+C to stop")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func startRabbitMQ(authService *auth.Service, cfg *config.Config) {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	server, err := mq.StartServer(authService, cfg.ServerID, rabbitURL)
	if err != nil {
		log.Fatalf("Failed to start RabbitMQ server: %v", err)
	}
	defer server.Close()

	log.Printf("RabbitMQ Server [%s] consuming queues %s / %s", cfg.ServerID, mq.LoginQueue, mq.ValidateQueue)
	log.Println("Test credentials: username=alice, password=password123")
	log.Println("Press Ctrl+C to stop")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("\nShutting down RabbitMQ server gracefully...")
}
