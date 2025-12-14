package handlers

import (
	"encoding/json"
	"log"
	"time"

	"github.com/aslotsu/notification-worker/models"
	"github.com/nats-io/nats.go"
)

type NotificationWorker struct {
	nats                *nats.Conn
	notificationService *NotificationService
	subscription        *nats.Subscription
}

// NewNotificationWorker creates a new notification worker
func NewNotificationWorker(nc *nats.Conn, notifService *NotificationService) *NotificationWorker {
	return &NotificationWorker{
		nats:                nc,
		notificationService: notifService,
	}
}

// Start begins listening for notification events
func (w *NotificationWorker) Start() error {
	log.Println("👂 Starting notification worker...")

	// Subscribe to all notification events using wildcard
	sub, err := w.nats.Subscribe("notifications.>", w.handleEvent)
	if err != nil {
		return err
	}

	w.subscription = sub
	log.Println("✅ Subscribed to notifications.> (all notification events)")

	return nil
}

// Stop gracefully stops the worker
func (w *NotificationWorker) Stop() error {
	log.Println("🛑 Stopping notification worker...")

	if w.subscription != nil {
		if err := w.subscription.Unsubscribe(); err != nil {
			return err
		}
	}

	return nil
}

// handleEvent processes a notification event from NATS
func (w *NotificationWorker) handleEvent(msg *nats.Msg) {
	startTime := time.Now()

	log.Printf("📨 Received event on subject: %s", msg.Subject)

	// Parse event
	var event models.NotificationEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("❌ Failed to unmarshal event: %v", err)
		log.Printf("   Raw data: %s", string(msg.Data))
		return
	}

	// Validate event
	if err := w.validateEvent(&event); err != nil {
		log.Printf("❌ Invalid event: %v", err)
		return
	}

	// Skip if user is triggering action on their own content
	if event.Owner == event.TriggerUser {
		log.Printf("⏭️  Skipping self-notification: owner=%s, trigger=%s", event.Owner, event.TriggerUser)
		return
	}

	// Convert event to notification
	notification := models.Notification{
		Owner:        event.Owner,
		UserId:       event.TriggerUser,
		UserName:     event.Username,
		UserPic:      event.UserPicture,
		UserBio:      event.UserBio,
		Action:       event.Action,
		ResourceType: event.ResourceType,
		ResourceId:   event.ResourceID,
		Excerpt:      event.Excerpt,
		CreatedAt:    time.Unix(event.CreatedAt, 0),
	}

	// Create notification in DynamoDB
	if err := w.notificationService.CreateNotification(notification); err != nil {
		log.Printf("❌ Failed to create notification: %v", err)

		// TODO: Implement retry logic or dead letter queue
		// For now, just log the error
		return
	}

	duration := time.Since(startTime)
	log.Printf("✅ Processed notification in %v (owner=%s, action=%d, resource=%s)",
		duration, event.Owner, event.Action, event.ResourceID)
}

// validateEvent validates the notification event
func (w *NotificationWorker) validateEvent(event *models.NotificationEvent) error {
	if event.Owner == "" {
		return &ValidationError{Field: "owner", Message: "owner is required"}
	}

	if event.TriggerUser == "" {
		return &ValidationError{Field: "trigger_user", Message: "trigger_user is required"}
	}

	if event.Action == 0 {
		return &ValidationError{Field: "action", Message: "action is required"}
	}

	if event.ResourceID == "" {
		return &ValidationError{Field: "resource_id", Message: "resource_id is required"}
	}

	if event.ResourceType == "" {
		return &ValidationError{Field: "resource_type", Message: "resource_type is required"}
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
