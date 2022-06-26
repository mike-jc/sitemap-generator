package writers

import (
	"encoding/xml"
	"io"
	"sitemap-generator/pkg/writers/models"
)

type SitemapWriter interface {
	Write(data models.Sitemap) error
}

type sitemapWriter struct {
	dest io.Writer
}

func NewSitemapWriter(dest io.Writer) SitemapWriter {
	return &sitemapWriter{
		dest: dest,
	}
}

func (sw *sitemapWriter) Write(data models.Sitemap) error {
	bytes, err := xml.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	bytes = []byte(xml.Header + string(bytes))
	if _, err = sw.dest.Write(bytes); err != nil {
		return err
	}
	return nil
}
