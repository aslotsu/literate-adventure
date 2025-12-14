package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type TestEvent struct {
	Owner        string `json:"owner"`
	TriggerUser  string `json:"trigger_user"`
	Username     string `json:"username"`
	UserPicture  string `json:"user_picture"`
	UserBio      string `json:"user_bio"`
	Action       int    `json:"action"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Excerpt      string `json:"excerpt"`
	CreatedAt    int64  `json:"created_at"`
}

func main() {
	log.Println("📤 Publishing test notification event...")

	// Connect to NATS
	nc, err := nats.Connect(
		"nats://connect.ngs.global",
		nats.UserCredentials("NGS-Default-exobook.creds"),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer nc.Close()

	log.Println("✅ Connected to NATS")

	// Create test event
	event := TestEvent{
		Owner:        "test-user-123",
		TriggerUser:  "test-user-456",
		Username:     "John Doe",
		UserPicture:  "https://example.com/pic.jpg",
		UserBio:      "Test user bio",
		Action:       1, // Like post
		ResourceType: "POST",
		ResourceID:   "test-post-789",
		Excerpt:      "This is a test notification",
		CreatedAt:    time.Now().Unix(),
	}

	// Marshal to JSON
	eventData, err := json.Marshal(event)
	if err != nil {
		log.Fatalf("Failed to marshal event: %v", err)
	}

	// Publish event
	subject := "notifications.post.like"
	if err := nc.Publish(subject, eventData); err != nil {
		log.Fatalf("Failed to publish: %v", err)
	}

	log.Printf("✅ Published event to %s", subject)
	log.Printf("   Owner: %s", event.Owner)
	log.Printf("   Trigger User: %s (%s)", event.Username, event.TriggerUser)
	log.Printf("   Resource: %s (%s)", event.ResourceType, event.ResourceID)

	// Wait for message to be delivered
	time.Sleep(2 * time.Second)

	log.Println("🎉 Test event published successfully!")
	log.Println("👀 Check notification-worker logs to see if it was processed")
}
