package photo

import "time"

const DefaultTimelineLimit = 50

type TimelineItem struct {
	ID           string
	Name         string
	MimeType     string
	SizeBytes    int64
	CreatedAt    time.Time
	ObjectKey    string
	ThumbnailURL string
}

type TimelineGroup struct {
	Date  string
	Items []TimelineItem
}

type Timeline struct {
	Groups     []TimelineGroup
	NextBefore string
}

type Album struct {
	ID        string
	OwnerID   string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	ItemCount int
}

type AlbumDetail struct {
	Album
	Items []TimelineItem
}

type ThumbnailPrefs struct {
	ImageEnabled bool
	VideoEnabled bool
}
