package models

import "time"

type CrawlerContext struct {
	Location     string    `json:"location"`
	LastModified time.Time `json:"lastModified"`
	IsHtml       bool      `json:"isHtml"`
	Depth        int       `json:"depth"`
}

func (cc *CrawlerContext) Id() string {
	return cc.Location
}
