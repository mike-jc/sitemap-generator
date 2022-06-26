package models

import "time"

type UrlInfo struct {
	IsHtml       bool
	LastModified time.Time
}
