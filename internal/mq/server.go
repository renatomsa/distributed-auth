package mq

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/renatomsa/auth-grpc/internal/auth"
)

type Server struct {
	conn        *amqp.Connection
	channel     *amqp.Channel
	authService *auth.Service
	serverID    string
}

func StartServer(authService *auth.Service, serverID, url string) (*Server, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	if _, err := ch.QueueDeclare(LoginQueue, false, false, false, false, nil); err != nil {
		conn.Close()
		return nil, err
	}

	if _, err := ch.QueueDeclare(ValidateQueue, false, false, false, false, nil); err != nil {
		conn.Close()
		return nil, err
	}

	server := &Server{
		conn:        conn,
		channel:     ch,
		authService: authService,
		serverID:    serverID,
	}

	server.consumeQueues()
	return server, nil
}

func (s *Server) Close() {
	if s.channel != nil {
		_ = s.channel.Close()
	}
	if s.conn != nil {
		_ = s.conn.Close()
	}
}

func (s *Server) consumeQueues() {
	loginMsgs, err := s.channel.Consume(LoginQueue, "", true, false, false, false, nil)
	if err != nil {
		log.Printf("[RabbitMQ %s] Failed to consume login queue: %v", s.serverID, err)
	}

	validateMsgs, err := s.channel.Consume(ValidateQueue, "", true, false, false, false, nil)
	if err != nil {
		log.Printf("[RabbitMQ %s] Failed to consume validate queue: %v", s.serverID, err)
	}

	go s.handleLogin(loginMsgs)
	go s.handleValidate(validateMsgs)
}

func (s *Server) handleLogin(msgs <-chan amqp.Delivery) {
	for d := range msgs {
		var req AuthRequest
		if err := json.Unmarshal(d.Body, &req); err != nil {
			log.Printf("[RabbitMQ %s] Invalid login payload: %v", s.serverID, err)
			continue
		}

		log.Printf("[RabbitMQ %s] Handling login for user=%s corr_id=%s", s.serverID, req.Username, d.CorrelationId)
		resp, err := s.authService.Authenticate(req.Username, req.Password)
		if err != nil {
			log.Printf("[RabbitMQ %s] Login error: %v", s.serverID, err)
			continue
		}

		s.publishResponse(d, AuthResponse{
			Token:   resp.Token,
			Success: resp.Success,
			Message: resp.Message,
		})
	}
}

func (s *Server) handleValidate(msgs <-chan amqp.Delivery) {
	for d := range msgs {
		var req AuthRequest
		if err := json.Unmarshal(d.Body, &req); err != nil {
			log.Printf("[RabbitMQ %s] Invalid validate payload: %v", s.serverID, err)
			continue
		}

		log.Printf("[RabbitMQ %s] Handling validate corr_id=%s", s.serverID, d.CorrelationId)
		resp, err := s.authService.ValidateToken(req.Token)
		if err != nil {
			log.Printf("[RabbitMQ %s] Validate error: %v", s.serverID, err)
			continue
		}

		s.publishResponse(d, AuthResponse{
			Valid:    resp.Valid,
			UserID:   resp.UserID,
			Username: resp.Username,
			Message:  resp.Message,
		})
	}
}

func (s *Server) publishResponse(d amqp.Delivery, resp AuthResponse) {
	if d.ReplyTo == "" {
		log.Printf("[RabbitMQ %s] No replyTo set, skipping response", s.serverID)
		return
	}

	body, err := json.Marshal(resp)
	if err != nil {
		log.Printf("[RabbitMQ %s] Failed to marshal response: %v", s.serverID, err)
		return
	}

	err = s.channel.Publish(
		"",
		d.ReplyTo,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			Body:          body,
			CorrelationId: d.CorrelationId,
		},
	)
	if err != nil {
		log.Printf("[RabbitMQ %s] Failed to publish response: %v", s.serverID, err)
		return
	}
	log.Printf("[RabbitMQ %s] Responded corr_id=%s to %s", s.serverID, d.CorrelationId, d.ReplyTo)
}
