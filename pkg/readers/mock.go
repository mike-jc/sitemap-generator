package readers

import "sitemap-generator/pkg/readers/models"

type ReaderMockOptions struct {
	CheckUrl func(url string) (info models.UrlInfo, err error)
	ReadUrl  func(url string) (body []byte, err error)
}

type readerMock struct {
	checkUrl func(url string) (info models.UrlInfo, err error)
	readUrl  func(url string) (body []byte, err error)
}

func NewReaderMock(opts ReaderMockOptions) Reader {
	return &readerMock{
		checkUrl: opts.CheckUrl,
		readUrl:  opts.ReadUrl,
	}
}

func (rm *readerMock) CheckUrl(url string) (info models.UrlInfo, err error) {
	return rm.checkUrl(url)
}

func (rm *readerMock) ReadUrl(url string) (body []byte, err error) {
	return rm.readUrl(url)
}
