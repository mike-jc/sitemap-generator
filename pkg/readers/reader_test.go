package readers_test

import (
	"net/http"
	"net/http/httptest"
	"sitemap-generator/pkg/readers"
	"sitemap-generator/utils"
	"testing"
	"time"
)

func TestReader_CheckUrl(t *testing.T) {
	lastModified, _ := time.Parse(time.RFC3339, "2022-05-11T12:48:18Z")
	reader := readers.NewReader(readers.ReaderOptions{})

	t.Run("is html and no last modified date", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
		}))
		defer srv.Close()

		info, err := reader.CheckUrl(srv.URL)
		utils.AssertNoError(t, err)
		utils.AssertEmpty(t, info.LastModified)
		utils.AssertTrue(t, info.IsHtml)
	})

	t.Run("not html and has last modified date", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Last-Modified", lastModified.Format(time.RFC1123))
		}))
		defer srv.Close()

		info, err := reader.CheckUrl(srv.URL)
		utils.AssertNoError(t, err)
		utils.AssertEqual(t, info.LastModified, lastModified)
		utils.AssertFalse(t, info.IsHtml)
	})
}

func TestReader_ReadUrl(t *testing.T) {
	expected := []byte("Hello, world!")
	reader := readers.NewReader(readers.ReaderOptions{
		Timeout:      200 * time.Millisecond,
		MaxRetries:   3,
		MaxRedirects: 3,
	})

	t.Run("successfully read", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(expected)
		}))
		defer srv.Close()

		body, err := reader.ReadUrl(srv.URL)
		utils.AssertNoError(t, err)
		utils.AssertEqual(t, string(body), string(expected))
	})

	t.Run("timeout while reading", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(400 * time.Millisecond)
		}))
		defer srv.Close()

		_, err := reader.ReadUrl(srv.URL)
		utils.AssertTrue(t, readers.IsTimeout(err))
	})

	t.Run("too many redirects", func(t *testing.T) {
		srv := httptest.NewServer(http.RedirectHandler("/", http.StatusMovedPermanently))
		defer srv.Close()

		_, err := reader.ReadUrl(srv.URL)
		utils.AssertTrue(t, readers.IsTooManyRedirects(err))
	})

	t.Run("too many retries", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusMovedPermanently)
			// no Location header will produce error in HTTP client
		}))
		defer srv.Close()

		_, err := reader.ReadUrl(srv.URL)
		utils.AssertHasError(t, err, "Maximum retries exceeded with error")
	})
}
