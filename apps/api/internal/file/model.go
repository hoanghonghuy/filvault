package file

import "time"

const (
	StatusPending = "PENDING"
	StatusReady   = "READY"
	StatusFailed  = "FAILED"
)

type File struct {
	ID              string
	OwnerID         string
	FolderID        *string
	Name            string
	OriginalName    string
	ObjectKey       string
	MimeType        string
	SizeBytes       int64
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
	UploadExpiresAt *time.Time
}

func ObjectKey(ownerID, fileID string) string {
	return "users/" + ownerID + "/files/" + fileID
}
