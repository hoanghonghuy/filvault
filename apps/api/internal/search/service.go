package search

import (
	"context"
	"strings"
	"time"

	"filnest/internal/apperr"
)

const DefaultLimit = 50

type FolderHit struct {
	ID        string
	ParentID  *string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FileHit struct {
	ID        string
	Name      string
	MimeType  string
	SizeBytes int64
	UpdatedAt time.Time
}

type Result struct {
	Folders []FolderHit
	Files   []FileHit
}

type Repository interface {
	SearchFolders(ctx context.Context, ownerID, pattern string, limit int) ([]FolderHit, error)
	SearchFiles(ctx context.Context, ownerID, pattern string, limit int) ([]FileHit, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Search(ctx context.Context, ownerID, q string, limit int) (Result, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return Result{}, apperr.Validation
	}
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > 100 {
		limit = 100
	}
	pattern := ilikePattern(q)
	folders, err := s.repo.SearchFolders(ctx, ownerID, pattern, limit)
	if err != nil {
		return Result{}, err
	}
	files, err := s.repo.SearchFiles(ctx, ownerID, pattern, limit)
	if err != nil {
		return Result{}, err
	}
	return Result{Folders: folders, Files: files}, nil
}

func ilikePattern(q string) string {
	q = strings.ReplaceAll(q, `\`, `\\`)
	q = strings.ReplaceAll(q, `%`, `\%`)
	q = strings.ReplaceAll(q, `_`, `\_`)
	return q
}
