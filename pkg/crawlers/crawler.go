package crawlers

import (
	"fmt"
	"sitemap-generator/pkg/crawlers/models"
	"sitemap-generator/pkg/parsers"
	"sitemap-generator/pkg/readers"
	"sitemap-generator/pkg/workerPools"
	"sitemap-generator/pkg/writers"
	writersModels "sitemap-generator/pkg/writers/models"
	"sitemap-generator/services"
	"sitemap-generator/utils"
)

type CrawlerOptions struct {
	MaxDepth      int
	Logger        services.Logger
	WorkerPool    workerPools.WorkerPool
	Reader        readers.Reader
	Parser        parsers.Parser
	SitemapWriter writers.SitemapWriter
}

type Crawler interface {
	Traverse(startUrl string) ([]*models.Url, error)
	WriteSitemap(urls []*models.Url) error
}

type crawler struct {
	maxDepth int

	logger     services.Logger
	workerPool workerPools.WorkerPool

	reader readers.Reader
	parser parsers.Parser

	sitemapWriter writers.SitemapWriter
}

func NewCrawler(opts CrawlerOptions) Crawler {
	return &crawler{
		maxDepth:      opts.MaxDepth,
		logger:        opts.Logger,
		workerPool:    opts.WorkerPool,
		reader:        opts.Reader,
		parser:        opts.Parser,
		sitemapWriter: opts.SitemapWriter,
	}
}

func (c *crawler) Traverse(startUrl string) ([]*models.Url, error) {
	urls := make([]*models.Url, 0)

	// get links from the start URL
	c.logger.Debug("Crawler: starting initial scan", startUrl)
	result, err := c.scanUrlForLinks(&models.CrawlerContext{
		Location: startUrl,
	})
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		c.logger.Debug("Crawler: no links extracted by following the start URL")
		return urls, nil
	}
	c.logger.Debug("Crawler: scanned the start URL", utils.InJSON(result))

	// handler processes URL extracting all links from the page
	var handler workerPools.WorkerHandler = func(v interface{}) error {
		ctx, ok := v.(*models.CrawlerContext)
		if !ok {
			return fmt.Errorf("Crawler: unknown type of input: %T", v)
		}

		c.logger.Debug("Crawler: starting to scan URL", startUrl)
		if result, err := c.scanUrlForLinks(ctx); err != nil {
			return err
		} else {
			c.logger.Debug("Crawler: scanned URL. Dispatching", utils.InJSON(result))
			for _, r := range result {
				c.workerPool.Dispatch(r, c.needsTask)
			}
			return nil
		}
	}

	// start workers and add all links extracted from the start URL to the task queue
	if _, err := c.workerPool.Init(handler); err != nil {
		return nil, fmt.Errorf("Crawler: could not initialize worker pool: %s", err.Error())
	}
	c.logger.Debug("Crawler: worker pool initialized")

	// collect initial results and put necessary tasks to the queue
	for _, r := range result {
		c.workerPool.Dispatch(r, c.needsTask)
	}

	// wait until all links extracted or max depth is reached
	c.workerPool.WaitFinalize()
	c.logger.Debug("Crawler: tasks completed")

	// convert to necessary result type
	results, _ := c.workerPool.Results()
	for _, r := range results {
		if ctx, ok := r.(*models.CrawlerContext); ok {
			urls = append(urls, &models.Url{
				Location:     ctx.Location,
				LastModified: ctx.LastModified,
			})
		} else {
			c.logger.Warn(fmt.Sprintf("Crawler: wrong type of worker pool's result: %T", r))
		}
	}

	return urls, nil
}

func (c *crawler) scanUrlForLinks(ctx *models.CrawlerContext) ([]*models.CrawlerContext, error) {
	result := make([]*models.CrawlerContext, 0)

	c.logger.Debug("Crawler: starting to read URL", ctx)
	body, err := c.reader.ReadUrl(ctx.Location)
	if err != nil {
		c.logger.Warn("Crawler: could not read URL", ctx.Location, err.Error())
		return result, err
	}
	c.logger.Debug(fmt.Sprintf("Crawler: got body (length: %d bytes)", len(body)))

	c.logger.Debug("Crawler: starting to parse HTML")
	urls := c.parser.ParseHtmlForLinks(ctx.Location, body)
	c.logger.Debug("Crawler: got links", urls)

	urls = utils.StringSliceUnique(urls)
	for _, u := range urls {
		// such check could be duplicated by other workers if they meet this URL on pages they scan,
		// but it's a cheap price to avoid a waiting for the end of a slow or timed-out check by ALL workers
		c.logger.Debug("Crawler: checking if URL acceptable", u)
		urlInfo, err := c.reader.CheckUrl(u)

		if err != nil {
			c.logger.Warn("Crawler: could not read URL while checking", u, err.Error())
		} else {
			uCtx := &models.CrawlerContext{
				Location:     u,
				LastModified: urlInfo.LastModified,
				IsHtml:       urlInfo.IsHtml,
				Depth:        ctx.Depth + 1,
			}
			result = append(result, uCtx)
			c.logger.Debug("Crawler: checked URL", uCtx)
		}
	}
	return result, nil
}

func (c *crawler) needsTask(v interface{}) bool {
	ctx, ok := v.(*models.CrawlerContext)
	if !ok {
		c.logger.Error(fmt.Sprintf("Crawler: unknown type of input while checking if result needs a task to be processed: %T", v))
		return false
	}

	c.logger.Debug("Crawler: checking if result needs a task to be processed", ctx)
	if ctx.IsHtml {
		c.logger.Debug("Crawler: URL is HTML page", ctx)
		if ctx.Depth < c.maxDepth {
			return true
		} else {
			c.logger.Info(fmt.Sprintf("Crawler: skip scanning, would be too deep for max depth %d", c.maxDepth), ctx)
			return false
		}
	} else {
		c.logger.Debug("Crawler: not HTML page, skip it", ctx)
		return false
	}
}

func (c *crawler) WriteSitemap(urls []*models.Url) error {
	data := writersModels.Sitemap{}

	data.Urls = make([]writersModels.SiteUrl, len(urls))
	for i, u := range urls {
		data.Urls[i] = writersModels.BuildSitemapUrl(u.Location, u.LastModified)
	}

	c.logger.Debug("Crawler: prepared sitemap data", data.Urls)
	return c.sitemapWriter.Write(data)
}
