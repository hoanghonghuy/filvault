package file

import (
	"testing"

	"filvault/internal/apperr"
	"filvault/internal/platform/config"
)

func TestValidateUpload_AllImageTypes(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		contentType string
		size        int64
		wantErr     error
	}{
		// Standard JPEG
		{"JPG standard", "photo.jpg", "image/jpeg", 1024, nil},
		{"JPEG standard", "photo.jpeg", "image/jpeg", 2048, nil},
		{"JPG uppercase", "PHOTO.JPG", "image/jpeg", 1024, nil},
		{"JPEG uppercase", "IMAGE.JPEG", "IMAGE/JPEG", 2048, nil},
		{"JPG progressive", "photo.jpg", "image/pjpeg", 1024, nil},
		{"JPG with charset", "photo.jpg", "image/jpeg; charset=utf-8", 1024, nil},

		// Standard PNG
		{"PNG standard", "graphic.png", "image/png", 512, nil},
		{"PNG uppercase", "GRAPHIC.PNG", "image/png", 512, nil},
		{"PNG x-png", "graphic.png", "image/x-png", 512, nil},

		// GIF
		{"GIF standard", "animation.gif", "image/gif", 4096, nil},
		{"GIF uppercase", "ANIMATION.GIF", "image/gif", 4096, nil},

		// WebP
		{"WebP standard", "modern.webp", "image/webp", 8192, nil},
		{"WebP uppercase", "MODERN.WEBP", "image/webp", 8192, nil},

		// HEIC / HEIF
		{"HEIC standard", "camera.heic", "image/heic", 3000, nil},
		{"HEIC uppercase", "IMG_4921.HEIC", "image/heic", 3000, nil},
		{"HEIF standard", "camera.heif", "image/heif", 3000, nil},
		{"HEIF uppercase", "IMG_4921.HEIF", "image/heif", 3000, nil},
		{"HEIC sequence", "burst.heic", "image/heic-sequence", 5000, nil},
		{"HEIF sequence", "burst.heif", "image/heif-sequence", 5000, nil},
		{"HEIC x-heic Android", "samsung.heic", "image/x-heic", 3000, nil},
		{"HEIF x-heif Android", "samsung.heif", "image/x-heif", 3000, nil},
		{"HEIC with params", "camera.heic", "image/heic; param=value", 3000, nil},

		// Edge cases: Size
		{"Zero byte image", "empty.jpg", "image/jpeg", 0, nil},
		{"Max allowed size", "large.png", "image/png", config.MaxFileSizeBytes, nil},
		{"Too large image", "toolarge.heic", "image/heic", config.MaxFileSizeBytes + 1, apperr.FileTooLarge},
		{"Negative size", "bad.jpg", "image/jpeg", -1, apperr.Validation},

		// Disallowed / Invalid types
		{"SVG rejected", "vector.svg", "image/svg+xml", 1024, apperr.Validation},
		{"EXE rejected", "malware.exe", "application/x-msdownload", 1024, apperr.Validation},
		{"Mismatched extension and MIME", "photo.jpg", "image/png", 1024, apperr.Validation},
		{"HEIC mismatched MIME", "photo.heic", "image/png", 1024, apperr.Validation},
		{"Empty filename", "", "image/jpeg", 1024, apperr.Validation},
		{"Empty contentType", "photo.jpg", "", 1024, apperr.Validation},
		{"No extension", "photo", "image/jpeg", 1024, apperr.Validation},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateUpload(tc.filename, tc.contentType, tc.size)
			if err != tc.wantErr {
				t.Fatalf("ValidateUpload(%q, %q, %d) = %v; want %v", tc.filename, tc.contentType, tc.size, err, tc.wantErr)
			}
		})
	}
}
