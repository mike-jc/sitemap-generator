package utils

import (
	"net/url"
)

// UrlPercentEncode encodes any non-ASCII character to percent-encoded (%C3%BC)
// Use it instead of standard url.QueryEscape() if you don't need
// to encode URL reserved characters (/, :, ?, & etc)
func UrlPercentEncode(v string) string {
	u, err := url.Parse(v)
	if err != nil {
		return v
	}
	return u.String()
}
