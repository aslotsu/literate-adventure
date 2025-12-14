package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/aslotsu/notification-worker/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/pusher/pusher-http-go/v5"
)

type NotificationService struct {
	client       *dynamodb.Client
	tableName    string
	pusherClient *pusher.Client
}

// NewNotificationService creates a new notification service
func NewNotificationService(region, tableName, pusherAppID, pusherKey, pusherSecret, pusherCluster string) (*NotificationService, error) {
	// Load AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %v", err)
	}

	// Create DynamoDB client
	client := dynamodb.NewFromConfig(cfg)

	// Initialize Pusher client for real-time notifications
	var pusherClient *pusher.Client
	if pusherAppID != "" && pusherKey != "" && pusherSecret != "" {
		pusherClient = &pusher.Client{
			AppID:   pusherAppID,
			Key:     pusherKey,
			Secret:  pusherSecret,
			Cluster: pusherCluster,
			Secure:  true,
		}
		log.Println("✅ Pusher client initialized for real-time notifications")
	} else {
		log.Println("⚠️ Pusher credentials not provided - real-time notifications disabled")
	}

	return &NotificationService{
		client:       client,
		tableName:    tableName,
		pusherClient: pusherClient,
	}, nil
}

// CreateNotification creates a notification in DynamoDB
func (s *NotificationService) CreateNotification(notif models.Notification) error {
	// Generate unique ID
	notif.Id = uuid.New().String()

	// Generate action key for deduplication
	notif.ActionKey = notif.GenerateActionKey()

	// Set read status to false by default
	notif.ReadStatus = false

	// Marshal to DynamoDB format
	item, err := attributevalue.MarshalMap(notif)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %v", err)
	}

	// Check if notification already exists (deduplication)
	// We use action_key to prevent duplicate notifications
	// For example: user likes same post multiple times, only create one notification
	existing, err := s.client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		IndexName:              aws.String("OwnerIndex"),
		KeyConditionExpression: aws.String("#owner = :owner"),
		FilterExpression:       aws.String("action_key = :action_key"),
		ExpressionAttributeNames: map[string]string{
			"#owner": "owner", // 'owner' is a reserved keyword
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":owner":      &types.AttributeValueMemberS{Value: notif.Owner},
			":action_key": &types.AttributeValueMemberS{Value: notif.ActionKey},
		},
		Limit: aws.Int32(1),
	})

	if err != nil {
		log.Printf("Warning: Failed to check for duplicate notification: %v", err)
		// Continue anyway - better to have duplicate than miss notification
	} else if existing.Count > 0 {
		log.Printf("Notification already exists for action_key: %s, skipping", notif.ActionKey)
		return nil // Not an error, just skip
	}

	// Store in DynamoDB
	_, err = s.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
	})

	if err != nil {
		return fmt.Errorf("failed to create notification: %v", err)
	}

	log.Printf("✅ Created notification: owner=%s, action=%d, resource=%s",
		notif.Owner, notif.Action, notif.ResourceId)

	// Trigger Pusher event for real-time notification delivery
	if s.pusherClient != nil {
		go s.triggerPusherNotification(notif)
	}

	return nil
}

// triggerPusherNotification sends a real-time notification via Pusher
func (s *NotificationService) triggerPusherNotification(notif models.Notification) {
	channelName := fmt.Sprintf("user-%s-notifications", notif.Owner)

	// Create event data matching frontend expectations
	eventData := map[string]interface{}{
		"id":           notif.Id,
		"action":       notif.Action,
		"username":     notif.UserName,
		"user_id":      notif.UserId,
		"user_pic":     notif.UserPic,
		"resource_id":  notif.ResourceId,
		"resource_type": notif.ResourceType,
		"excerpt":      notif.Excerpt,
		"read_status":  notif.ReadStatus,
		"created_at":   notif.CreatedAt,
		"action_key":   notif.ActionKey,
	}

	err := s.pusherClient.Trigger(channelName, "new-notification", eventData)
	if err != nil {
		log.Printf("❌ Failed to trigger Pusher notification for user %s: %v", notif.Owner, err)
	} else {
		log.Printf("📤 Pusher notification sent to channel: %s", channelName)
	}
}

// GetNotificationsByOwner retrieves notifications for a user
func (s *NotificationService) GetNotificationsByOwner(owner string, limit int32) ([]models.Notification, error) {
	resp, err := s.client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		IndexName:              aws.String("OwnerIndex"),
		KeyConditionExpression: aws.String("owner = :owner"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":owner": &types.AttributeValueMemberS{Value: owner},
		},
		ScanIndexForward: aws.Bool(false), // Latest first
		Limit:            aws.Int32(limit),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %v", err)
	}

	var notifications []models.Notification
	err = attributevalue.UnmarshalListOfMaps(resp.Items, &notifications)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal notifications: %v", err)
	}

	return notifications, nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(owner, actionKey string) error {
	_, err := s.client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"owner":      &types.AttributeValueMemberS{Value: owner},
			"action_key": &types.AttributeValueMemberS{Value: actionKey},
		},
		UpdateExpression: aws.String("SET #read = :true"),
		ExpressionAttributeNames: map[string]string{
			"#read": "read",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":true": &types.AttributeValueMemberBOOL{Value: true},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %v", err)
	}

	return nil
}
