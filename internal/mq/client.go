package mq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Client struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	replyQueue string
	replies    <-chan amqp.Delivery
}

func NewClient(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	reply, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		conn.Close()
		return nil, err
	}

	msgs, err := ch.Consume(reply.Name, "", true, true, false, false, nil)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{
		conn:       conn,
		channel:    ch,
		replyQueue: reply.Name,
		replies:    msgs,
	}, nil
}

func (c *Client) Close() {
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *Client) Login(username, password string) (*AuthResponse, error) {
	return c.call(LoginQueue, AuthRequest{Username: username, Password: password})
}

func (c *Client) Validate(token string) (*AuthResponse, error) {
	return c.call(ValidateQueue, AuthRequest{Token: token})
}

func (c *Client) call(queue string, payload AuthRequest) (*AuthResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	corrID := fmt.Sprintf("corr-%d", rand.Int63())
	err = c.channel.Publish(
		"",
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			Body:          body,
			ReplyTo:       c.replyQueue,
			CorrelationId: corrID,
		},
	)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		select {
		case msg := <-c.replies:
			if msg.CorrelationId != corrID {
				continue
			}

			var resp AuthResponse
			if err := json.Unmarshal(msg.Body, &resp); err != nil {
				return nil, err
			}
			return &resp, nil
		case <-ctx.Done():
			return nil, errors.New("timeout waiting for RabbitMQ response")
		}
	}
}
