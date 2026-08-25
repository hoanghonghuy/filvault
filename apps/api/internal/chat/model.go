package chat

import "time"

type Conversation struct {
	ID        string
	OwnerID   string
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Message struct {
	ID             string
	ConversationID string
	OwnerID        string
	Body           string
	CreatedAt      time.Time
	Attachments    []Attachment
}

type Attachment struct {
	ID           string
	MessageID    string
	FileID       string
	OriginalName string
	Name         string
	MimeType     string
	SizeBytes    int64
	CreatedAt    time.Time
}

type UploadSession struct {
	FileID    string
	UploadURL string
	ExpiresAt time.Time
}
