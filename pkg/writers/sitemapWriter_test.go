package writers_test

import (
	"bytes"
	"sitemap-generator/pkg/writers"
	"sitemap-generator/pkg/writers/models"
	"sitemap-generator/utils"
	"testing"
)

func TestSitemapWriter_Write(t *testing.T) {
	data := models.Sitemap{
		Urls: []models.SiteUrl{
			{
				Location: "https://creativecommons.org/about/contact",
			},
			{
				Location:     "https://wiki.creativecommons.org/Intergovernmental_Organizations",
				LastModified: "2022-05-11T12:48:18Z",
			},
		},
	}

	expected := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://creativecommons.org/about/contact</loc>
  </url>
  <url>
    <loc>https://wiki.creativecommons.org/Intergovernmental_Organizations</loc>
    <lastmod>2022-05-11T12:48:18Z</lastmod>
  </url>
</urlset>`

	buffer := new(bytes.Buffer)
	sw := writers.NewSitemapWriter(buffer)

	err := sw.Write(data)
	utils.AssertNoError(t, err)
	utils.AssertEqual(t, buffer.String(), expected)
}
