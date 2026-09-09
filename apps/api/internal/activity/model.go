package activity

import "time"

// Event types recorded for the owner's activity feed (spec 09 §6).
const (
	TypeFileUploaded     = "file.uploaded"
	TypeFileTrashed      = "file.trashed"
	TypeFileRestored     = "file.restored"
	TypeFolderTrashed    = "folder.trashed"
	TypeFolderRestored   = "folder.restored"
	TypePurged           = "file.purged"
	TypeShareCreated     = "share.created"
	TypeShareRevoked     = "share.revoked"
	TypePasswordChanged  = "password.changed"
	TypeSettingsChanged  = "settings.changed"
)

// RetentionDays bounds how long events are kept before the ticker purges them.
const RetentionDays = 90

// DefaultListLimit matches the Settings feed default page size.
const DefaultListLimit = 20

// MaxListLimit caps a single page.
const MaxListLimit = 50

// Event is one recorded state change owned by a user.
type Event struct {
	ID         string    `json:"id"`
	OwnerID    string    `json:"-"`
	Type       string    `json:"type"`
	TargetName string    `json:"targetName"`
	CreatedAt  time.Time `json:"createdAt"`
}
