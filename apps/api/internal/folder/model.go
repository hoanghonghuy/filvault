package folder

import "time"

type Folder struct {
	ID        string
	OwnerID   string
	ParentID  *string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (f Folder) Alive() bool {
	return f.DeletedAt == nil
}

type BrowserFile struct {
	ID        string
	Name      string
	MimeType  string
	SizeBytes int64
	UpdatedAt time.Time
}

type Browser struct {
	Folder     *Folder
	Breadcrumb []Folder
	Folders    []Folder
	Files      []BrowserFile
}
