package objectstore

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrObjectNotFound = errors.New("object not found")

// Memory is an in-process ObjectStore for tests.
type Memory struct {
	mu      sync.Mutex
	objects map[string]ObjectStat
}

func NewMemory() *Memory {
	return &Memory{objects: map[string]ObjectStat{}}
}

func (m *Memory) CreateUploadURL(_ context.Context, key string, opts UploadOptions) (PresignedURL, error) {
	exp := opts.Expires
	if exp <= 0 {
		exp = 15 * time.Minute
	}
	return PresignedURL{URL: "memory://upload/" + key, ExpiresAt: time.Now().UTC().Add(exp)}, nil
}

func (m *Memory) CreateDownloadURL(_ context.Context, key string, opts DownloadOptions) (PresignedURL, error) {
	exp := opts.Expires
	if exp <= 0 {
		exp = 5 * time.Minute
	}
	return PresignedURL{URL: "memory://download/" + key, ExpiresAt: time.Now().UTC().Add(exp)}, nil
}

func (m *Memory) Head(_ context.Context, key string) (ObjectStat, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	stat, ok := m.objects[key]
	if !ok {
		return ObjectStat{}, ErrObjectNotFound
	}
	return stat, nil
}

func (m *Memory) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.objects, key)
	return nil
}

// PutObject simulates a client PUT after presign.
func (m *Memory) PutObject(key string, stat ObjectStat) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.objects[key] = stat
}
