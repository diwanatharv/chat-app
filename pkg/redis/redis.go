package cache

import (
	"chat-app/pkg/config"
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type Observer interface {
	Notify(message string)
}

type SupportAgent struct {
	Name string
}

func (s *SupportAgent) Notify(message string) {
	fmt.Printf("Agent %s received notification: %s\n", s.Name, message)
}

type Customer struct {
	Name string
}

func (c *Customer) Notify(message string) {
	fmt.Printf("Customer %s received notification: %s\n", c.Name, message)
}

type ChatNotifier struct {
	observers []Observer
}

func (c *ChatNotifier) Attach(observer Observer) {
	c.observers = append(c.observers, observer)
}

func (c *ChatNotifier) NotifyAll(message string) {
	for _, observer := range c.observers {
		observer.Notify(message)
	}
}

// Strategy interface for handling different types of messages
type MessageHandler interface {
	HandleMessage(message string)
}

// Concrete Strategy - Text Message Handler
type TextMessageHandler struct{}

func (t *TextMessageHandler) HandleMessage(message string) {
	fmt.Println("Handling text message:", message)
}

// Concrete Strategy - Image Message Handler
type ImageMessageHandler struct{}

func (i *ImageMessageHandler) HandleMessage(message string) {
	fmt.Println("Handling image message:", message)
}

// Context for choosing a strategy dynamically
type MessageContext struct {
	handler MessageHandler
}

func (m *MessageContext) SetHandler(handler MessageHandler) {
	m.handler = handler
}

func (m *MessageContext) ProcessMessage(message string) {
	m.handler.HandleMessage(message)
}

// RedisClient defines the methods for interacting with Redis.
type RedisClient interface {
	PublishMessage(channel string, message string) error
	SubscribeToChannel(channel string, messageChannel chan<- string) error
	UnsubscribeFromChannel(channel string) error
}

// redisImpl is an implementation of RedisClient.
type redisImpl struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisClient() RedisClient {
	return &redisImpl{
		client: config.GetRedisClient(),
		ctx:    config.GetContext(),
	}
}

// PublishMessage publishes a message to a Redis channel.
func (r *redisImpl) PublishMessage(channel string, message string) error {
	return r.client.Publish(r.ctx, channel, message).Err()
}

// UnsubscribeFromChannel unsubscribes from a Redis channel.
func (r *redisImpl) UnsubscribeFromChannel(channel string) error {
	pubsub := r.client.Subscribe(r.ctx, channel)
	defer pubsub.Close()

	return pubsub.Unsubscribe(r.ctx, channel)
}

// SubscribeToChannel subscribes to a Redis channel and sends messages to the provided channel.

func (r *redisImpl) SubscribeToChannel(channel string, messageChannel chan<- string) error {
	pubsub := r.client.Subscribe(r.ctx, channel)
	defer pubsub.Close()

	for {
		select {
		case msg := <-pubsub.Channel():
			messageChannel <- msg.Payload
		case <-r.ctx.Done():
			return nil
		}
	}
}
