package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/renatomsa/auth-grpc/internal/mq"
	pb "github.com/renatomsa/auth-grpc/proto"
)

const (
	loadBalancerAddr = "localhost:9100"
	rabbitURL        = "amqp://guest:guest@localhost:5672/"
)

type loginFunc func(username, password string) (token string, success bool, message string, err error)
type validateFunc func(token string) (valid bool, userID int, username string, message string, err error)

func main() {
	transport := strings.ToLower(os.Getenv("TRANSPORT"))
	if transport == "" {
		transport = "grpc"
	}

	switch transport {
	case "rabbitmq":
		fmt.Println("Testing Authentication System via RabbitMQ")
		fmt.Println("=====================================")
		fmt.Printf("RabbitMQ target: %s\n", getRabbitURL())
		fmt.Println("-------------------------------------")

		if err := runRabbitTests(); err != nil {
			log.Printf("RabbitMQ test failed: %v\n", err)
			return
		}
	default:
		fmt.Println("Testing gRPC Authentication System via Load Balancer")
		fmt.Println("=====================================")
		fmt.Printf("Load Balancer target: %s\n", loadBalancerAddr)
		fmt.Println("-------------------------------------")

		if err := runGRPCTests(); err != nil {
			log.Printf("gRPC test failed: %v\n", err)
			return
		}
	}

	fmt.Println("All tests completed!")
}

func runGRPCTests() error {
	conn, err := grpc.NewClient(
		loadBalancerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)

	return runTests(
		func(username, password string) (string, bool, string, error) {
			resp, err := client.Login(context.Background(), &pb.LoginRequest{
				Username: username,
				Password: password,
			})
			if err != nil {
				return "", false, "", err
			}
			return resp.Token, resp.Success, resp.Message, nil
		},
		func(token string) (bool, int, string, string, error) {
			resp, err := client.ValidateToken(context.Background(), &pb.ValidateTokenRequest{
				Token: token,
			})
			if err != nil {
				return false, 0, "", "", err
			}
			return resp.Valid, int(resp.UserId), resp.Username, resp.Message, nil
		},
	)
}

func runRabbitTests() error {
	client, err := mq.NewClient(getRabbitURL())
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer client.Close()

	return runTests(
		func(username, password string) (string, bool, string, error) {
			resp, err := client.Login(username, password)
			if err != nil {
				return "", false, "", err
			}
			return resp.Token, resp.Success, resp.Message, nil
		},
		func(token string) (bool, int, string, string, error) {
			resp, err := client.Validate(token)
			if err != nil {
				return false, 0, "", "", err
			}
			return resp.Valid, resp.UserID, resp.Username, resp.Message, nil
		},
	)
}

func runTests(login loginFunc, validate validateFunc) error {
	fmt.Println("Test 1: Valid credentials (alice/password123)")
	token, success, message, err := login("alice", "password123")
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if success {
		fmt.Printf("Login successful\n")
		fmt.Printf("Token: %s...\n", shortenToken(token))

		fmt.Println("\nTest 2: Validate token")
		valid, userID, username, msg, err := validate(token)
		if err != nil {
			return fmt.Errorf("validate failed: %w", err)
		}

		if valid {
			fmt.Printf("Token is valid\n")
			fmt.Printf("User: %s (ID: %d)\n", username, userID)
		} else {
			fmt.Printf("Token is invalid: %s\n", msg)
		}
	} else {
		fmt.Printf("Login failed: %s\n", message)
	}

	fmt.Println("\nTest 3: Invalid credentials (alice/wrongpass)")
	_, success, _, err = login("alice", "wrongpass")
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if !success {
		fmt.Printf("Invalid credentials correctly rejected\n")
	} else {
		fmt.Printf("Invalid credentials were accepted (BUG!)\n")
	}

	fmt.Println("\nTest 4: Validate invalid token")
	valid, _, _, msg, err := validate("invalid.token.here")
	if err != nil {
		return fmt.Errorf("validate failed: %w", err)
	}

	if !valid {
		fmt.Printf("Invalid token correctly rejected\n")
	} else {
		fmt.Printf("Invalid token was accepted (BUG!) (%s)\n", msg)
	}

	return nil
}

func getRabbitURL() string {
	if val := os.Getenv("RABBITMQ_URL"); val != "" {
		return val
	}
	return rabbitURL
}

func shortenToken(token string) string {
	if len(token) > 50 {
		return token[:50]
	}
	return token
}
