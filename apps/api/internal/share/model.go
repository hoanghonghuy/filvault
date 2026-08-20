package share

import "time"

const (
	ResourceFile   = "file"
	ResourceFolder = "folder"
)

type Share struct {
	ID           string
	OwnerID      string
	ResourceType string
	ResourceID   string
	RecipientID  string
	CreatedAt    time.Time
}

// UserRef is the minimal user info exposed on share listings.
type UserRef struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

// Outgoing is a share the current user created.
type Outgoing struct {
	ID           string    `json:"id"`
	ResourceType string    `json:"resourceType"`
	ResourceID   string    `json:"resourceId"`
	ResourceName string    `json:"resourceName"`
	Recipient    UserRef   `json:"recipient"`
	CreatedAt    time.Time `json:"createdAt"`
}

// Incoming is a share the current user received.
type Incoming struct {
	ID           string    `json:"id"`
	ResourceType string    `json:"resourceType"`
	ResourceID   string    `json:"resourceId"`
	ResourceName string    `json:"resourceName"`
	Owner        UserRef   `json:"owner"`
	CreatedAt    time.Time `json:"createdAt"`
}
