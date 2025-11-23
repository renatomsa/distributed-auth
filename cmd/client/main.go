package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/renatomsa/auth-grpc/proto"
)

const loadBalancerAddr = "localhost:9100"

func main() {
	fmt.Println("Testing gRPC Authentication System via Load Balancer")
	fmt.Println("=====================================")

	fmt.Printf("Load Balancer target: %s\n", loadBalancerAddr)
	fmt.Println("-------------------------------------")

	if err := testServer(loadBalancerAddr); err != nil {
		log.Printf("Load balancer test failed: %v\n", err)
		return
	}

	fmt.Println("All tests completed!")
}

func testServer(addr string) error {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)

	fmt.Println("Test 1: Valid credentials (alice/password123)")
	loginResp, err := client.Login(context.Background(), &pb.LoginRequest{
		Username: "alice",
		Password: "password123",
	})
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if loginResp.Success {
		fmt.Printf("Login successful\n")
		fmt.Printf("Token: %s...\n", loginResp.Token[:50])
		token := loginResp.Token

		fmt.Println("\nTest 2: Validate token")
		validateResp, err := client.ValidateToken(context.Background(), &pb.ValidateTokenRequest{
			Token: token,
		})
		if err != nil {
			return fmt.Errorf("validate failed: %w", err)
		}

		if validateResp.Valid {
			fmt.Printf("Token is valid\n")
			fmt.Printf("User: %s (ID: %d)\n", validateResp.Username, validateResp.UserId)
		} else {
			fmt.Printf("Token is invalid: %s\n", validateResp.Message)
		}

	} else {
		fmt.Printf("Login failed: %s\n", loginResp.Message)
	}

	fmt.Println("\nTest 3: Invalid credentials (alice/wrongpass)")
	loginResp2, err := client.Login(context.Background(), &pb.LoginRequest{
		Username: "alice",
		Password: "wrongpass",
	})
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if !loginResp2.Success {
		fmt.Printf("Invalid credentials correctly rejected\n")
	} else {
		fmt.Printf("Invalid credentials were accepted (BUG!)\n")
	}

	fmt.Println("\nTest 4: Validate invalid token")
	validateResp2, err := client.ValidateToken(context.Background(), &pb.ValidateTokenRequest{
		Token: "invalid.token.here",
	})
	if err != nil {
		return fmt.Errorf("validate failed: %w", err)
	}

	if !validateResp2.Valid {
		fmt.Printf("Invalid token correctly rejected\n")
	} else {
		fmt.Printf("Invalid token was accepted (BUG!)\n")
	}

	return nil
}
