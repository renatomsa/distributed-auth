package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/renatomsa/auth-grpc/internal/mq"
	pb "github.com/renatomsa/auth-grpc/proto"
)

type runner func(ctx context.Context, username, password string) (bool, error)

func main() {
	var (
		transport = flag.String("transport", "grpc", "grpc or rabbitmq")
		samples   = flag.Int("samples", 1, "number of runs (sample size)")
		total     = flag.Int("n", 100000, "requests per run")
		concur    = flag.Int("c", 1, "concurrency (number of workers)")
		grpcAddr  = flag.String("grpc-addr", "localhost:9100", "gRPC target (load balancer)")
		rabbitURL = flag.String("rabbit-url", "", "RabbitMQ URL (default amqp://guest:guest@localhost:5672/)")
		user      = flag.String("user", "alice", "username")
		pass      = flag.String("pass", "password123", "password")
		invalidP  = flag.Float64("invalid", 0.5, "fraction of requests with invalid credentials (0.0-1.0)")
	)
	flag.Parse()

	mode := strings.ToLower(*transport)
	if *rabbitURL == "" {
		if val := os.Getenv("RABBITMQ_URL"); val != "" {
			*rabbitURL = val
		} else {
			*rabbitURL = "amqp://guest:guest@localhost:5672/"
		}
	}

	var makeRunner func() (runner, func())

	switch mode {
	case "grpc":
		makeRunner = func() (runner, func()) {
			conn, err := grpc.NewClient(
				*grpcAddr,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				log.Fatalf("gRPC connect failed: %v", err)
			}
			client := pb.NewAuthServiceClient(conn)
			run := func(ctx context.Context, username, password string) (bool, error) {
				resp, err := client.Login(ctx, &pb.LoginRequest{
					Username: username,
					Password: password,
				})
				if err != nil {
					return false, err
				}
				return resp.Success, nil
			}
			cleanup := func() { conn.Close() }
			return run, cleanup
		}
	case "rabbitmq":
		makeRunner = func() (runner, func()) {
			client, err := mq.NewClient(*rabbitURL)
			if err != nil {
				log.Fatalf("RabbitMQ connect failed: %v", err)
			}
			run := func(ctx context.Context, username, password string) (bool, error) {
				resp, err := client.Login(username, password)
				if err != nil {
					return false, err
				}
				return resp.Success, nil
			}
			cleanup := func() { client.Close() }
			return run, cleanup
		}
	default:
		log.Fatalf("invalid transport: %s (use grpc or rabbitmq)", mode)
	}

	fmt.Printf("Benchmarking %s | samples=%d | requests per run=%d | concurrency=%d | invalid share=%.2f\n", mode, *samples, *total, *concur, *invalidP)

	for s := 1; s <= *samples; s++ {
		start := time.Now()
		jobs := make(chan bool, *total)
		invalidThreshold := int(float64(*total) * *invalidP)
		for i := 0; i < *total; i++ {
			if i < invalidThreshold {
				jobs <- false
			} else {
				jobs <- true
			}
		}
		close(jobs)

		var wg sync.WaitGroup
		var mu sync.Mutex
		successAuth, invalidAuth, failures := 0, 0, 0

		for i := 0; i < *concur; i++ {
			run, cleanup := makeRunner()
			wg.Add(1)
			go func(run runner, cleanup func()) {
				defer wg.Done()
				defer cleanup()
				ctx := context.Background()
				for useValid := range jobs {
					ok, err := run(ctx, *user, pickPassword(*pass, useValid))
					if err != nil {
						mu.Lock()
						failures++
						mu.Unlock()
						continue
					}
					mu.Lock()
					if ok {
						successAuth++
					} else {
						invalidAuth++
					}
					mu.Unlock()
				}
			}(run, cleanup)
		}

		wg.Wait()

		elapsed := time.Since(start)
		fmt.Printf("[Run %d/%d] %s | ok=%d invalid=%d failures=%d | rate=%.2f req/s\n",
			s, *samples,
			elapsed,
			successAuth,
			invalidAuth,
			failures,
			float64(successAuth+invalidAuth+failures)/elapsed.Seconds())
	}
}

func pickPassword(validPass string, useValid bool) string {
	if useValid {
		return validPass
	}
	return "wrong-password"
}
