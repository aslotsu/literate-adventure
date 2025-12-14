package models

// NotificationEvent represents an event published to NATS
// that should trigger a notification creation
type NotificationEvent struct {
	Owner        string `json:"owner"`         // User who receives the notification
	TriggerUser  string `json:"trigger_user"`  // User who triggered the action
	Username     string `json:"username"`      // Trigger user's display name
	UserPicture  string `json:"user_picture"`  // Trigger user's profile picture
	UserBio      string `json:"user_bio"`      // Trigger user's bio
	Action       int    `json:"action"`        // Action type (1=like post, 2=like comment, etc.)
	ResourceType string `json:"resource_type"` // Type of resource (POST, COMMENT, etc.)
	ResourceID   string `json:"resource_id"`   // ID of the resource
	Excerpt      string `json:"excerpt"`       // Optional excerpt/preview text
	CreatedAt    int64  `json:"created_at"`    // Unix timestamp
}

// Action types
const (
	ActionLikePost     = 1
	ActionLikeComment  = 2
	ActionReplyPost    = 3
	ActionReplyComment = 4
	ActionMention      = 5
	ActionFollow       = 6
)

// Resource types
const (
	ResourceTypePost    = "POST"
	ResourceTypeComment = "COMMENT"
	ResourceTypeUser    = "USER"
)
