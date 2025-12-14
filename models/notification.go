package models

import (
	"fmt"
	"time"
)

// Notification represents a notification stored in DynamoDB
// Schema matches dynamodb-go-api/models/notif.go
type Notification struct {
	Id           string    `dynamodbav:"id" json:"id"`
	Owner        string    `dynamodbav:"owner" json:"owner"`               // User receiving notification
	UserId       string    `dynamodbav:"userid" json:"userid"`             // User who triggered action
	UserName     string    `dynamodbav:"username" json:"username"`         // Trigger user's name
	UserPic      string    `dynamodbav:"userpic" json:"userpic"`           // Trigger user's picture
	UserBio      string    `dynamodbav:"userbio" json:"userbio"`           // Trigger user's bio
	Action       int       `dynamodbav:"action" json:"action"`             // Action type
	ResourceType string    `dynamodbav:"resource_type" json:"resource_type"` // POST, COMMENT, etc.
	ResourceId   string    `dynamodbav:"resource_id" json:"resource_id"`   // ID of the resource
	Excerpt      string    `dynamodbav:"excerpt" json:"excerpt"`           // Optional preview text
	ActionKey    string    `dynamodbav:"action_key" json:"action_key"`     // Composite key for deduplication
	ReadStatus   bool      `dynamodbav:"read_status" json:"read_status"`   // Read status
	CreatedAt    time.Time `dynamodbav:"created_at" json:"created_at"`     // Time (stored as String in DynamoDB)
}

// GenerateActionKey creates a unique key for deduplication
// Format: {userid}#{resource_id}#{action}#{timestamp}
// The timestamp is always set to zero time for deduplication purposes
func (n *Notification) GenerateActionKey() string {
	return fmt.Sprintf("%s#%s#%d#%s", n.UserId, n.ResourceId, n.Action, "0001-01-01T00:00:00Z")
}
