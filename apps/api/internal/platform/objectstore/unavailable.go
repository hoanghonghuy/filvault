package objectstore

import (
	"context"
	"errors"
)

// Unavailable is an ObjectStore that fails readiness checks for tests.
type Unavailable struct {
	Err error
}

func NewUnavailable(err error) *Unavailable {
	if err == nil {
		err = errors.New("object store unavailable")
	}
	return &Unavailable{Err: err}
}

func (u *Unavailable) CheckReady(context.Context) error {
	return u.Err
}

func (u *Unavailable) CreateUploadURL(context.Context, string, UploadOptions) (PresignedURL, error) {
	return PresignedURL{}, u.Err
}

func (u *Unavailable) CreateDownloadURL(context.Context, string, DownloadOptions) (PresignedURL, error) {
	return PresignedURL{}, u.Err
}

func (u *Unavailable) Head(context.Context, string) (ObjectStat, error) {
	return ObjectStat{}, u.Err
}

func (u *Unavailable) Delete(context.Context, string) error {
	return u.Err
}
