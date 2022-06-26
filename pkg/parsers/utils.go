package parsers

import (
	"golang.org/x/net/html"
	"net/url"
)

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
