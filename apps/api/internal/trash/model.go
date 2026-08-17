package trash

import (
	"time"

	"filnest/internal/file"
	"filnest/internal/folder"
)

type Item struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Type      string     `json:"type"` // "file" | "folder"
	SizeBytes *int64     `json:"sizeBytes,omitempty"`
	MimeType  *string    `json:"mimeType,omitempty"`
	DeletedAt time.Time  `json:"deletedAt"`
}

type List struct {
	Folders []Item `json:"folders"`
	Files   []Item `json:"files"`
}

func fileItem(f file.File) Item {
	size := f.SizeBytes
	mime := f.MimeType
	return Item{
		ID:        f.ID,
		Name:      f.Name,
		Type:      "file",
		SizeBytes: &size,
		MimeType:  &mime,
		DeletedAt: f.DeletedAt.UTC(),
	}
}

func folderItem(f folder.Folder) Item {
	return Item{
		ID:        f.ID,
		Name:      f.Name,
		Type:      "folder",
		DeletedAt: f.DeletedAt.UTC(),
	}
}
