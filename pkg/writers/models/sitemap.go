package models

import (
	"encoding/xml"
	"sitemap-generator/utils"
	"time"
)

type Sitemap struct {
	XMLName xml.Name  `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
	Urls    []SiteUrl `xml:"url"`
}

type SiteUrl struct {
	Location     string `xml:"loc"`
	LastModified string `xml:"lastmod,omitempty"`
}

func BuildSitemapUrl(loc string, lastMod time.Time) SiteUrl {
	lastModified := ""
	if !lastMod.IsZero() {
		lastModified = lastMod.Format(time.RFC3339)
	}

	return SiteUrl{
		Location:     utils.UrlPercentEncode(loc),
		LastModified: lastModified,
	}
}
