package user

import "time"

type Public struct {
	ID                     string    `json:"id"`
	Email                  string    `json:"email"`
	DisplayName            string    `json:"displayName"`
	EmailVerified          bool      `json:"emailVerified"`
	StorageUsed            int64     `json:"storageUsed"`
	StorageQuota           int64     `json:"storageQuota"`
	TrashAutoDeleteEnabled bool      `json:"trashAutoDeleteEnabled"`
	TrashRetentionDays     int       `json:"trashRetentionDays"`
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
		TrashAutoDeleteEnabled: u.TrashAutoDeleteEnabled,
		TrashRetentionDays:     u.TrashRetentionDays,
		CreatedAt:              u.CreatedAt.UTC(),
	}
}
