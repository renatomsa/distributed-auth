package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/renatomsa/auth-grpc/internal/auth"
	"github.com/renatomsa/auth-grpc/internal/database"
	grpcserver "github.com/renatomsa/auth-grpc/internal/grpc"
	"github.com/renatomsa/auth-grpc/pkg/config"
)

func main() {
	// Carregar configurações
	cfg := config.Load()

	log.Println("Starting Auth gRPC Server...")
	cfg.Print()

	// Conectar ao banco de dados
	db, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✅ Connected to PostgreSQL")

	// Executar migrations
	if err := database.RunMigrations(db.GetDB()); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed database (apenas em desenvolvimento)
	if cfg.Environment == "development" {
		if err := database.SeedUsers(db.GetDB()); err != nil {
			log.Printf("Warning: Failed to seed database: %v", err)
		}
	}

	// Criar serviço de autenticação
	authService := auth.NewService(db)
	log.Println("✅ Auth service initialized")

	// Criar listener TCP para gRPC
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Criar servidor gRPC
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10 * 1024 * 1024), // 10MB
		grpc.MaxSendMsgSize(10 * 1024 * 1024), // 10MB
	)

	// Registrar serviço
	grpcserver.RegisterServer(grpcServer, authService, cfg.ServerID)
	log.Printf("✅ gRPC server registered on port %s", cfg.GRPCPort)

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh

		log.Println("\n🛑 Shutting down gracefully...")
		grpcServer.GracefulStop()
		log.Println("✅ Server stopped")
	}()

	// Iniciar servidor
	log.Printf("🚀 gRPC Server [%s] listening on port %s", cfg.ServerID, cfg.GRPCPort)
	log.Println("📝 Test credentials: username=alice, password=password123")
	log.Println("Press Ctrl+C to stop")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
