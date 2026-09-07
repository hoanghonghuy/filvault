package user

import "time"

type Public struct {
	ID                     string    `json:"id"`
	Email                  string    `json:"email"`
	DisplayName            string    `json:"displayName"`
	EmailVerified          bool      `json:"emailVerified"`
	StorageUsed            int64     `json:"storageUsed"`
	StorageQuota           int64     `json:"storageQuota"`
	ImageThumbnailsEnabled bool      `json:"imageThumbnailsEnabled"`
	VideoThumbnailsEnabled bool      `json:"videoThumbnailsEnabled"`
	TrashAutoDeleteEnabled bool      `json:"trashAutoDeleteEnabled"`
	TrashRetentionDays     int       `json:"trashRetentionDays"`
	ActiveStatusEnabled    bool      `json:"activeStatusEnabled"`
	CreatedAt              time.Time `json:"createdAt"`
}

func PublicFrom(u User) Public {
	return Public{
		ID:                     u.ID,
		Email:                  u.Email,
		DisplayName:            u.DisplayName,
		EmailVerified:          u.EmailVerified(),
		StorageUsed:            u.StorageUsed,
		StorageQuota:           u.StorageQuota,
		ImageThumbnailsEnabled: u.ImageThumbnailsEnabled,
		VideoThumbnailsEnabled: u.VideoThumbnailsEnabled,
		TrashAutoDeleteEnabled: u.TrashAutoDeleteEnabled,
		TrashRetentionDays:     u.TrashRetentionDays,
		ActiveStatusEnabled:    u.ActiveStatusEnabled,
		CreatedAt:              u.CreatedAt.UTC(),
	}
}
