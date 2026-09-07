package chat

import "time"

const (
	AttachmentAvailable = "available"
	AttachmentTrashed   = "trashed"
	AttachmentPurged    = "purged"
)

type Conversation struct {
	ID          string
	OwnerID     string
	Title       string
	Type        string
	PeerID      string
	PeerName    string
	PeerEmail   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastMessage time.Time
}

type ConversationPreview struct {
	MessageID     string
	Body          string
	CreatedAt     time.Time
	AttachmentIDs []string
}

type ReadState struct {
	LastReadAt            *time.Time
	LastReadMessageID     string
	PeerLastReadAt        *time.Time
	PeerLastReadMessageID string
	UnreadCount           int
}

type Message struct {
	ID              string
	ConversationID  string
	OwnerID         string
	SenderID        string
	ClientMessageID string
	Body            string
	CreatedAt       time.Time
	EditedAt        *time.Time
	RemovedAt       *time.Time
	Idempotent      bool
	Attachments     []Attachment
}

type Event struct {
	ID             string
	Sequence       int64
	ConversationID string
	Type           string
	AggregateID    string
	Payload        []byte
	CreatedAt      time.Time
}

type EventCursor struct {
	Sequence int64
}

type Attachment struct {
	ID           string
	MessageID    string
	FileID       *string
	OriginalName string
	Name         string
	MimeType     string
	SizeBytes    int64
	CreatedAt    time.Time
	ThumbnailURL string
	Availability string
}

type UploadSession struct {
	FileID    string
	UploadURL string
	ExpiresAt time.Time
}
