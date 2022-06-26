package readers

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sitemap-generator/pkg/readers/models"
	"strings"
	"time"
)

type ReaderOptions struct {
	Timeout      time.Duration
	MaxRetries   int
	MaxRedirects int
}

type Reader interface {
	CheckUrl(url string) (info models.UrlInfo, err error)
	ReadUrl(url string) (body []byte, err error)
}

type reader struct {
	maxRetries int

	client http.Client
}

func NewReader(opts ReaderOptions) Reader {
	client := http.Client{
		Timeout: opts.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= opts.MaxRedirects {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}
	return &reader{
		maxRetries: opts.MaxRetries,
		client:     client,
	}
}

func (r *reader) CheckUrl(url string) (info models.UrlInfo, err error) {
	var resp *http.Response

	resp, err = r.doOrRetry(http.MethodHead, url, nil)
	if err != nil {
		return
	}

	// Is it HTML ?
	contentType := strings.Split(resp.Header.Get("Content-Type"), ";")[0]
	if contentType == "text/html" {
		info.IsHtml = true
	}

	// Try to get last time of resource modification
	lastModifiedHeader := resp.Header.Get("Last-Modified")
	if lastModifiedHeader != "" {
		info.LastModified, _ = time.Parse(time.RFC1123, lastModifiedHeader)
	}
	return
}

func (r *reader) ReadUrl(url string) (body []byte, err error) {
	var resp *http.Response
	resp, err = r.doOrRetry(http.MethodGet, url, nil)

	// connection error
	if err != nil {
		return nil, err
	}
	// http error
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error [%d] %s", resp.StatusCode, resp.Status)
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (r *reader) doOrRetry(method string, url string, reqBody io.Reader) (resp *http.Response, err error) {
	var req *http.Request
	req, err = http.NewRequest(method, url, reqBody)
	if err != nil {
		return
	}

	attempt := 1
	for {
		resp, err = r.client.Do(req)
		if err == nil {
			return
		}
		if IsTimeout(err) || IsTooManyRedirects(err) {
			return
		}
		if attempt < r.maxRetries {
			attempt++
		} else {
			err = fmt.Errorf("Maximum retries exceeded with error: %s", err.Error())
			return
		}
	}
}

func IsTimeout(err error) bool {
	return strings.Contains(err.Error(), "context deadline exceeded")
}

func IsTooManyRedirects(err error) bool {
	return strings.Contains(err.Error(), "too many redirects")
}
