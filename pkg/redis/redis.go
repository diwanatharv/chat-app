package cache

import (
	"chat-app/pkg/config"
	"log"
)

// PublishMessage publishes a message to the chat_room
func PublishMessage(channel string, message string) error {
	err := config.RedisClient.Publish(config.Ctx, channel, message).Err()
	if err != nil {
		log.Printf("Could not publish message to Redis: %v", err)
		return err
	}
	return nil
}

// SubscribeToChannel subscribes to a Redis channel
func SubscribeToChannel(channel string, handleMessage func(string)) {
	subscriber := config.RedisClient.Subscribe(config.Ctx, channel)

	go func() {
		for msg := range subscriber.Channel() {
			// Call the handler function to process the message
			handleMessage(msg.Payload)
		}
	}()
}
