package config

import (
	"net/url"
	"os"
	"strings"
)

func stringsTrim(s string) string  { return strings.TrimSpace(s) }
func stringsToLower(s string) string { return strings.ToLower(s) }
func stringsContains(s, sub string) bool { return strings.Contains(s, sub) }
func stringsHasPrefix(s, prefix string) bool { return strings.HasPrefix(s, prefix) }
func stringsEqualFold(a, b string) bool { return strings.EqualFold(a, b) }
func osGetenv(key string) string { return os.Getenv(key) }
func urlParse(raw string) (*url.URL, error) { return url.Parse(raw) }
