package photo

import "strings"

var photoMIMEs = map[string]struct{}{
	"image/jpeg":      {},
	"image/png":       {},
	"image/gif":       {},
	"image/webp":      {},
	"image/heic":      {},
	"image/heif":      {},
	"video/mp4":       {},
	"video/quicktime": {},
	"video/webm":      {},
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
