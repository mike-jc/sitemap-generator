package parsers_test

import (
	"sitemap-generator/pkg/parsers"
	"sitemap-generator/utils"
	"testing"
)

func TestParser_ParseHtmlForLinks(t *testing.T) {
	url := "https://example.com/home"
	parser := parsers.NewParser()

	t.Run("no base and page has relatives urls", func(t *testing.T) {
		body := `<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en">
<head>
    <title>sitemaps.org - Home</title>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
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

		expected := []string{
			"https://example.com/faq.php",
			"https://example.com/protocol.php",
			"https://example.com/home",
			"https://example.com/terms.php",
		}

		links := parser.ParseHtmlForLinks(url, []byte(body))
		utils.AssertEqual(t, links, expected)
	})

	t.Run("has base and page has relative urls", func(t *testing.T) {
		body := `<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en">
<head>
    <title>sitemaps.org - Home</title>
    <base href="https://www.w3schools.com/" target="_blank">
</head>
<body>
    <ul>
        <li><a href="faq.php">FAQ</a></li>
        <li><a href="protocol.php">Protocol</a></li>
    </ul>
</body>
</html>`

		expected := []string{
			"https://www.w3schools.com/faq.php",
			"https://www.w3schools.com/protocol.php",
		}

		links := parser.ParseHtmlForLinks(url, []byte(body))
		utils.AssertEqual(t, links, expected)
	})

	t.Run("no base and page has absolute and relatives urls", func(t *testing.T) {
		body := `<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en">
<head>
    <title>sitemaps.org - Home</title>
</head>
<body>
    <ul>
        <li><a href="faq.php">FAQ</a></li>
        <li><a href="https://www.w3schools.com/protocol.php">Protocol</a></li>
    </ul>
</body>
</html>`

		expected := []string{
			"https://example.com/faq.php",
			"https://www.w3schools.com/protocol.php",
		}

		links := parser.ParseHtmlForLinks(url, []byte(body))
		utils.AssertEqual(t, links, expected)
	})
}
