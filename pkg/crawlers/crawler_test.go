package crawlers_test

import (
	"os"
	"sitemap-generator/pkg/crawlers"
	"sitemap-generator/pkg/crawlers/models"
	"sitemap-generator/pkg/parsers"
	"sitemap-generator/pkg/readers"
	readersModels "sitemap-generator/pkg/readers/models"
	"sitemap-generator/pkg/workerPools"
	"sitemap-generator/services"
	"sitemap-generator/utils"
	"testing"
)

func TestCrawler_Traverse(t *testing.T) {
	startUrl := "https://my-example.com"
	body := `<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en">
<head>
    <title>sitemaps.org - Home</title>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
    <base href="https://my-example.com/" target="_blank">
</head>
<body>
    <div id="container">
        <div id="intro">
            <div id="pageHeader">
                <h1>sitemaps.org</h1>
            </div>
            <div id="selectionbar">
                <ul>
                    <li><a href="faq.php">FAQ</a></li>
                    <li><a href="protocol.php">Protocol</a></li>
                    <li class="activelink"><a href="#">Home</a></li>
                </ul>
            </div>
            <!-- end selectionbar -->
        </div>
        <!-- end intro -->
        <div style="padding: 14px; float: right;" id="languagebox">
            Language: <span id="langContainer"></span>
        </div>
        <div id="mainContent">
            <h1>What are Sitemaps?</h1>
            <p>
                Sitemaps are an easy way for webmasters to inform search engines about pages on
                their sites that are available for crawling. In its simplest form, a Sitemap is
                an XML file that lists URLs for a site along with additional metadata about each
                URL
            </p>
        </div>
        <!-- end maincontent -->
    </div>
    <!-- closes #container -->
    <div id="footer">
        <p><a href="terms.php">Terms and conditions</a></p>
    </div>
</body>
</html>`

	expectedUrls := []*models.Url{
		{
			Location: "https://my-example.com/",
		},
		{
			Location: "https://my-example.com/faq.php",
		},
		{
			Location: "https://my-example.com/protocol.php",
		},
		{
			Location: "https://my-example.com/terms.php",
		},
	}

	logger, err := services.NewLogger(os.Stderr, "testing", "error")
	utils.AssertNoError(t, err)

	wp := workerPools.NewWorkerPool(logger, 2)
	parser := parsers.NewParser()

	reader := readers.NewReaderMock(readers.ReaderMockOptions{
		CheckUrl: func(url string) (readersModels.UrlInfo, error) {
			return readersModels.UrlInfo{}, nil
		},
		ReadUrl: func(url string) ([]byte, error) {
			return []byte(body), nil
		},
	})

	c := crawlers.NewCrawler(crawlers.CrawlerOptions{
		MaxDepth:   1,
		Logger:     logger,
		WorkerPool: wp,
		Reader:     reader,
		Parser:     parser,
	})

	urls, err := c.Traverse(startUrl)
	utils.AssertNoError(t, err)
	utils.AssertEqualSlices(t, urls, expectedUrls)
}
