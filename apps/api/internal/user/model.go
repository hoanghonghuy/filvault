package user

import "time"

type User struct {
	ID                     string
	Email                  string
	DisplayName            string
	PasswordHash           string
	EmailVerifiedAt        *time.Time
	VerificationCodeHash   string
	VerificationExpiresAt  *time.Time
	StorageUsed            int64
	StorageQuota           int64
	ImageThumbnailsEnabled bool
	VideoThumbnailsEnabled bool
	TrashAutoDeleteEnabled bool
	TrashRetentionDays     int
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

func (u User) EmailVerified() bool {
	return u.EmailVerifiedAt != nil
}
