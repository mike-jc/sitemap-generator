package parsers

import (
	"bytes"
	"golang.org/x/net/html"
	"net/url"
)

type Parser interface {
	ParseHtmlForLinks(bodyUrl string, body []byte) []string
}

type parser struct {
}

func NewParser() Parser {
	return &parser{}
}

// ParseHtmlForLinks parses HTML doc to find all <A> tags and extract URL (taking into account the <base> tag)
func (p *parser) ParseHtmlForLinks(bodyUrl string, body []byte) []string {
	urls := make([]string, 0)
	tokenizer := html.NewTokenizer(bytes.NewReader(body))

	base := parseUrlWithoutFragment(bodyUrl)
	for {
		next := tokenizer.Next()

		// error or end of the HTML doc
		if next == html.ErrorToken {
			return urls
		}

		// HTML tag appeared
		if next == html.StartTagToken || next == html.SelfClosingTagToken {
			token := tokenizer.Token()
			switch token.Data {
			case "base":
				if href := tokenAttrByKey(token, "href"); href != "" {
					base = parseUrlWithoutFragment(href)
				}
			case "a":
				if href := tokenAttrByKey(token, "href"); href != "" {
					if a := parseUrlWithoutFragment(href); a != nil {
						if !a.IsAbs() && base != nil {
							a = base.ResolveReference(a)
						}
						if a.Host != "" {
							if a.Scheme == "" {
								a.Scheme = "http"
							}
							urls = append(urls, a.String())
						}
					}
				}
			}
		}
	}
}

func parseUrlWithoutFragment(v string) *url.URL {
	if u, err := url.Parse(v); err == nil {
		u.Fragment = ""
		return u
	} else {
		return nil
	}
}

func tokenAttrByKey(token html.Token, key string) string {
	for _, a := range token.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}
