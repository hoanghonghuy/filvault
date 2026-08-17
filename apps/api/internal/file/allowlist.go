package file

import (
	"path/filepath"
	"strings"

	"filnest/internal/apperr"
	"filnest/internal/platform/config"
)

var allowlist = map[string][]string{
	"jpg":  {"image/jpeg"},
	"jpeg": {"image/jpeg"},
	"png":  {"image/png"},
	"gif":  {"image/gif"},
	"webp": {"image/webp"},
	"heic": {"image/heic", "image/heif"},
	"heif": {"image/heic", "image/heif"},
	"mp4":  {"video/mp4"},
	"mov":  {"video/quicktime"},
	"webm": {"video/webm"},
	"pdf":  {"application/pdf"},
	"doc":  {"application/msword"},
	"docx": {"application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
	"xls":  {"application/vnd.ms-excel"},
	"xlsx": {"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
	"ppt":  {"application/vnd.ms-powerpoint"},
	"pptx": {"application/vnd.openxmlformats-officedocument.presentationml.presentation"},
	"txt":  {"text/plain"},
	"md":   {"text/markdown"},
	"csv":  {"text/csv"},
	"zip":  {"application/zip"},
}

func ValidateUpload(name, contentType string, size int64) error {
	name = strings.TrimSpace(name)
	contentType = strings.TrimSpace(contentType)
	if name == "" || contentType == "" || size < 0 {
		return apperr.Validation
	}
	if size > config.MaxFileSizeBytes {
		return apperr.FileTooLarge
	}
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))
	if ext == "" {
		return apperr.Validation
	}
	allowed, ok := allowlist[ext]
	if !ok {
		return apperr.Validation
	}
	for _, mime := range allowed {
		if mime == contentType {
			return nil
		}
	}
	return apperr.Validation
}
