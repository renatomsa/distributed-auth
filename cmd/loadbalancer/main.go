package main

import (
	"context"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "github.com/renatomsa/auth-grpc/proto"
)

var defaultBackends = []string{
	"localhost:9001",
	"localhost:9002",
	"localhost:9003",
}

const loadBalancerPort = "9100"

type loadBalancer struct {
	pb.UnimplementedAuthServiceServer
	backends []string
}

func (lb *loadBalancer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	backend, conn, err := lb.connectToBackend()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	log.Printf("[LoadBalancer] Forwarding Login to %s", backend)
	client := pb.NewAuthServiceClient(conn)
	return client.Login(ctx, req)
}

func (lb *loadBalancer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	backend, conn, err := lb.connectToBackend()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	log.Printf("[LoadBalancer] Forwarding ValidateToken to %s", backend)
	client := pb.NewAuthServiceClient(conn)
	return client.ValidateToken(ctx, req)
}

func (lb *loadBalancer) connectToBackend() (string, *grpc.ClientConn, error) {
	backend := lb.backends[rand.Intn(len(lb.backends))]

	conn, err := grpc.NewClient(
		backend,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("[LoadBalancer] Failed to connect to backend %s: %v", backend, err)
		return "", nil, status.Errorf(codes.Unavailable, "backend %s unavailable", backend)
	}

	return backend, conn, nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	backends := defaultBackends
	if len(backends) == 0 {
		log.Fatal("no backend servers configured")
	}

	lis, err := net.Listen("tcp", ":"+loadBalancerPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", loadBalancerPort, err)
	}

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024),
		grpc.MaxSendMsgSize(10*1024*1024),
	)

	pb.RegisterAuthServiceServer(grpcServer, &loadBalancer{backends: backends})

	log.Printf("Load balancer listening on port %s", loadBalancerPort)
	log.Printf("Backend servers: %v", backends)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh

		log.Println("\nShutting down load balancer gracefully...")
		grpcServer.GracefulStop()
		log.Println("Load balancer stopped")
	}()

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve load balancer: %v", err)
	}
}
