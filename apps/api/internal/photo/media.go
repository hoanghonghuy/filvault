package photo

import "strings"

var photoMIMEs = map[string]struct{}{
	"image/jpeg":          {},
	"image/pjpeg":         {},
	"image/png":           {},
	"image/x-png":         {},
	"image/gif":           {},
	"image/webp":          {},
	"image/heic":          {},
	"image/heif":          {},
	"image/heic-sequence": {},
	"image/heif-sequence": {},
	"image/x-heic":        {},
	"image/x-heif":        {},
	"video/mp4":           {},
	"video/quicktime":     {},
	"video/webm":          {},
}

func IsPhotoMime(mime string) bool {
	_, ok := photoMIMEs[strings.TrimSpace(mime)]
	return ok
}

func PhotoMIMESlice() []string {
	out := make([]string, 0, len(photoMIMEs))
	for m := range photoMIMEs {
		out = append(out, m)
	}
	return out
}
