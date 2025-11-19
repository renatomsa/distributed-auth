package grpc

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/renatomsa/auth-grpc/internal/auth"
	pb "github.com/renatomsa/auth-grpc/proto"
)

type Server struct {
	pb.UnimplementedAuthServiceServer
	authService *auth.Service
	serverID    string
}

func NewServer(authService *auth.Service, serverID string) *Server {
	return &Server{
		authService: authService,
		serverID:    serverID,
	}
}

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Printf("[gRPC Server %s] Received Login request for user: %s", s.serverID, req.Username)

	if req.Username == "" || req.Password == "" {
		log.Printf("[gRPC Server %s] Invalid input: username or password empty", s.serverID)
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	resp, err := s.authService.Authenticate(req.Username, req.Password)
	if err != nil {
		log.Printf("[gRPC Server %s] Authentication error: %v", s.serverID, err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	log.Printf("[gRPC Server %s] Login response: success=%v", s.serverID, resp.Success)

	return &pb.LoginResponse{
		Token:   resp.Token,
		Success: resp.Success,
		Message: resp.Message,
	}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	log.Printf("[gRPC Server %s] Received ValidateToken request", s.serverID)

	if req.Token == "" {
		log.Printf("[gRPC Server %s] Invalid input: token empty", s.serverID)
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	resp, err := s.authService.ValidateToken(req.Token)
	if err != nil {
		log.Printf("[gRPC Server %s] Validation error: %v", s.serverID, err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	log.Printf("[gRPC Server %s] Token validation response: valid=%v", s.serverID, resp.Valid)

	return &pb.ValidateTokenResponse{
		Valid:    resp.Valid,
		UserId:   int32(resp.UserID),
		Username: resp.Username,
		Message:  resp.Message,
	}, nil
}

func (s *Server) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.LoginResponse, error) {
	log.Printf("[gRPC Server %s] Received RefreshToken request", s.serverID)

	if req.Token == "" {
		log.Printf("[gRPC Server %s] Invalid input: token empty", s.serverID)
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	resp, err := s.authService.RefreshToken(req.Token)
	if err != nil {
		log.Printf("[gRPC Server %s] Refresh error: %v", s.serverID, err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	log.Printf("[gRPC Server %s] Token refresh response: success=%v", s.serverID, resp.Success)

	return &pb.LoginResponse{
		Token:   resp.Token,
		Success: resp.Success,
		Message: resp.Message,
	}, nil
}

func RegisterServer(grpcServer *grpc.Server, authService *auth.Service, serverID string) {
	server := NewServer(authService, serverID)
	pb.RegisterAuthServiceServer(grpcServer, server)
}
