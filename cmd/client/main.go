package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/renatomsa/auth-grpc/proto"
)

// Servidores disponíveis
var servers = []string{
	"localhost:9001",
	"localhost:9002",
	"localhost:9003",
}

func main() {
	fmt.Println("🧪 Testing gRPC Authentication System")
	fmt.Println("=====================================\n")

	// Testar cada servidor
	for i, addr := range servers {
		fmt.Printf("📡 Testing Server %d: %s\n", i+1, addr)
		fmt.Println("-------------------------------------")

		if err := testServer(addr); err != nil {
			log.Printf("❌ Server %d failed: %v\n", i+1, err)
		}

		fmt.Println()
	}

	fmt.Println("✨ All tests completed!")
}

func testServer(addr string) error {
	// Conectar ao servidor gRPC
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)

	// Teste 1: Login com credenciais válidas
	fmt.Println("Test 1: Valid credentials (alice/password123)")
	loginResp, err := client.Login(context.Background(), &pb.LoginRequest{
		Username: "alice",
		Password: "password123",
	})
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if loginResp.Success {
		fmt.Printf("✅ Login successful\n")
		fmt.Printf("🎫 Token: %s...\n", loginResp.Token[:50])
		token := loginResp.Token

		// Teste 2: Validar token
		fmt.Println("\nTest 2: Validate token")
		validateResp, err := client.ValidateToken(context.Background(), &pb.ValidateTokenRequest{
			Token: token,
		})
		if err != nil {
			return fmt.Errorf("validate failed: %w", err)
		}

		if validateResp.Valid {
			fmt.Printf("✅ Token is valid\n")
			fmt.Printf("👤 User: %s (ID: %d)\n", validateResp.Username, validateResp.UserId)
		} else {
			fmt.Printf("❌ Token is invalid: %s\n", validateResp.Message)
		}

		// Teste 3: Refresh token
		fmt.Println("\nTest 3: Refresh token")
		refreshResp, err := client.RefreshToken(context.Background(), &pb.RefreshTokenRequest{
			Token: token,
		})
		if err != nil {
			return fmt.Errorf("refresh failed: %w", err)
		}

		if refreshResp.Success {
			fmt.Printf("✅ Token refreshed\n")
			fmt.Printf("🎫 New Token: %s...\n", refreshResp.Token[:50])
		} else {
			fmt.Printf("❌ Refresh failed: %s\n", refreshResp.Message)
		}

	} else {
		fmt.Printf("❌ Login failed: %s\n", loginResp.Message)
	}

	// Teste 4: Login com credenciais inválidas
	fmt.Println("\nTest 4: Invalid credentials (alice/wrongpass)")
	loginResp2, err := client.Login(context.Background(), &pb.LoginRequest{
		Username: "alice",
		Password: "wrongpass",
	})
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if !loginResp2.Success {
		fmt.Printf("✅ Invalid credentials correctly rejected\n")
	} else {
		fmt.Printf("❌ Invalid credentials were accepted (BUG!)\n")
	}

	// Teste 5: Validar token inválido
	fmt.Println("\nTest 5: Validate invalid token")
	validateResp2, err := client.ValidateToken(context.Background(), &pb.ValidateTokenRequest{
		Token: "invalid.token.here",
	})
	if err != nil {
		return fmt.Errorf("validate failed: %w", err)
	}

	if !validateResp2.Valid {
		fmt.Printf("✅ Invalid token correctly rejected\n")
	} else {
		fmt.Printf("❌ Invalid token was accepted (BUG!)\n")
	}

	return nil
}
