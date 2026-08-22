package search

import (
	"context"
	"strings"
	"time"

	"filvault/internal/apperr"
)

const (
	DefaultLimit = 50
	MaxLimit     = 50
	MaxQueryLen  = 200
)

type TypeFilter string

const (
	TypeAll      TypeFilter = "all"
	TypeImage    TypeFilter = "image"
	TypeVideo    TypeFilter = "video"
	TypeDocument TypeFilter = "document"
	TypeArchive  TypeFilter = "archive"
	TypeFolder   TypeFilter = "folder"
)

func (t TypeFilter) includesFolders() bool {
	return t == TypeAll || t == TypeFolder
}

func ParseType(raw string) (TypeFilter, error) {
	if raw == "" {
		return TypeAll, nil
	}
	switch t := TypeFilter(raw); t {
	case TypeAll, TypeImage, TypeVideo, TypeDocument, TypeArchive, TypeFolder:
		return t, nil
	default:
		return "", apperr.Validation
	}
}

type SortField string

const (
	SortRelevance SortField = "relevance"
	SortName      SortField = "name"
	SortDate      SortField = "date"
	SortSize      SortField = "size"
)

func ParseSort(raw string) (SortField, error) {
	if raw == "" {
		return SortRelevance, nil
	}
	switch s := SortField(raw); s {
	case SortRelevance, SortName, SortDate, SortSize:
		return s, nil
	default:
		return "", apperr.Validation
	}
}

type Order int

const (
	OrderDesc Order = iota
	OrderAsc
)

func ParseOrder(raw string) (Order, error) {
	if raw == "" {
		return OrderDesc, nil
	}
	switch raw {
	case "desc":
		return OrderDesc, nil
	case "asc":
		return OrderAsc, nil
	default:
		return 0, apperr.Validation
	}
}

// Query carries the validated search contract of spec 09 S1. All fields are
// normalized; zero values mean "not set" except type/sort/order which default.
type Query struct {
	Text     string
	Type     TypeFilter
	FolderID *string
	From     *time.Time
	To       *time.Time
	Sort     SortField
	Order    Order
	Limit    int
}

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
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Result struct {
	Folders []FolderHit
	Files   []FileHit
}

type Repository interface {
	SearchFolders(ctx context.Context, ownerID string, q Query) ([]FolderHit, error)
	SearchFiles(ctx context.Context, ownerID string, q Query) ([]FileHit, error)
	FolderExists(ctx context.Context, ownerID, folderID string) (bool, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Search(ctx context.Context, ownerID string, q Query) (Result, error) {
	q.Text = strings.TrimSpace(q.Text)
	if q.Text == "" || len(q.Text) > MaxQueryLen {
		return Result{}, apperr.Validation
	}
	if q.Limit <= 0 {
		q.Limit = DefaultLimit
	}
	if q.Limit > MaxLimit {
		return Result{}, apperr.Validation
	}
	if (q.From != nil && q.To != nil) && q.From.After(*q.To) {
		return Result{}, apperr.Validation
	}
	if q.FolderID != nil {
		exists, err := s.repo.FolderExists(ctx, ownerID, *q.FolderID)
		if err != nil {
			return Result{}, err
		}
		if !exists {
			return Result{}, apperr.NotFound
		}
	}
	folders := []FolderHit{}
	if q.Type.includesFolders() {
		var err error
		folders, err = s.repo.SearchFolders(ctx, ownerID, q)
		if err != nil {
			return Result{}, err
		}
	}
	files := []FileHit{}
	if q.includesFiles() {
		var err error
		files, err = s.repo.SearchFiles(ctx, ownerID, q)
		if err != nil {
			return Result{}, err
		}
	}
	return Result{Folders: folders, Files: files}, nil
}

// includesFiles reports whether the filter asks for file hits at all. Only
// type=folder excludes files; every mime-based type maps to a file subset.
func (q Query) includesFiles() bool { return q.Type != TypeFolder }

// ParseDateOnly parses an ISO date (YYYY-MM-DD) or returns nil when empty.
func ParseDateOnly(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse(time.DateOnly, raw)
	if err != nil {
		return nil, apperr.Validation
	}
	return &t, nil
}
