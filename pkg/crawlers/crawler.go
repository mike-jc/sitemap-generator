package crawlers

import (
	"fmt"
	"sitemap-generator/pkg/crawlers/models"
	"sitemap-generator/pkg/parsers"
	"sitemap-generator/pkg/readers"
	"sitemap-generator/pkg/workerPools"
	"sitemap-generator/services"
	"sitemap-generator/utils"
	"sync"
)

type CrawlerOptions struct {
	MaxDepth   int
	Logger     services.Logger
	Reader     readers.Reader
	Parser     parsers.Parser
	WorkerPool workerPools.WorkerPool
}

type Crawler interface {
	Traverse(startUrl string) ([]*models.Url, error)
}

type crawler struct {
	maxDepth int

	logger     services.Logger
	reader     readers.Reader
	parser     parsers.Parser
	workerPool workerPools.WorkerPool

	resultsLocker sync.Mutex
	urls          map[string]*models.Url
}

func NewCrawler(opts CrawlerOptions) Crawler {
	return &crawler{
		maxDepth:   opts.MaxDepth,
		logger:     opts.Logger,
		reader:     opts.Reader,
		parser:     opts.Parser,
		workerPool: opts.WorkerPool,
	}
}

func (c *crawler) Traverse(startUrl string) ([]*models.Url, error) {
	c.urls = make(map[string]*models.Url)

	// a wrap to cast input to the needed type
	var handler workerPools.WorkerHandler = func(v interface{}) error {
		ctx, ok := v.(models.CrawlerContext)
		if !ok {
			return fmt.Errorf("Crawler: unknown type of input in worker handler: %T", v)
		}
		return c.traverseIteration(ctx)
	}

	if _, err := c.workerPool.Init(handler); err != nil {
		return nil, fmt.Errorf("Crawler: could not initialize worker pool: %s", err.Error())
	}
	c.logger.Debug("Crawler: worker pool initialized")

	// analyze the start URL, collect links and put initial tasks to the queue
	err := c.traverseIteration(models.CrawlerContext{
		Location: startUrl,
	})
	if err != nil {
		return nil, err
	}

	// wait until all links extracted or max depth is reached
	c.workerPool.WaitFinalize()
	c.logger.Debug("Crawler: tasks completed")

	// convert to necessary result type
	results := make([]*models.Url, 0)
	for _, u := range c.urls {
		results = append(results, u)
	}

	return results, nil
}

func (c *crawler) traverseIteration(ctx models.CrawlerContext) error {
	c.logger.Debug("Crawler: starting to scan URL", ctx)
	result, err := c.scanUrlForLinks(ctx)
	if err != nil {
		return err
	}
	c.logger.Debug("Crawler: URL scanned", utils.InJSON(result))

	// add to result and produce new task if needed
	for _, r := range result {
		added := c.addResult(&models.Url{
			Location:     r.Location,
			LastModified: r.LastModified,
		})
		if added {
			c.dispatch(r)
		}
	}
	c.logger.Debug("Crawler: scan result dispatched")
	return nil
}

func (c *crawler) scanUrlForLinks(ctx models.CrawlerContext) ([]models.CrawlerContext, error) {
	result := make([]models.CrawlerContext, 0)

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

		if err == nil {
			uCtx := models.CrawlerContext{
				Location:     u,
				LastModified: urlInfo.LastModified,
				IsHtml:       urlInfo.IsHtml,
				Depth:        ctx.Depth + 1,
			}
			result = append(result, uCtx)
			c.logger.Debug("Crawler: checked URL", uCtx)
		} else {
			c.logger.Warn("Crawler: could not read URL while checking", u, err.Error())
		}
	}
	return result, nil
}

// dispatch adds a task to the queue if needed
func (c *crawler) dispatch(ctx models.CrawlerContext) {
	if ctx.IsHtml {
		c.logger.Debug("Crawler: URL is HTML page", ctx)
		if ctx.Depth < c.maxDepth {
			c.workerPool.AddTask(ctx)
		} else {
			c.logger.Info(fmt.Sprintf("Crawler: skip scanning, would be too deep for max depth %d", c.maxDepth), ctx)
			return
		}
	} else {
		c.logger.Debug("Crawler: not HTML page, skip it", ctx)
		return
	}
}

// addResult collects URL if it's not yet in the list and returns *true*;
// URL list is being locked while reading from and writing in
func (c *crawler) addResult(url *models.Url) bool {
	c.resultsLocker.Lock()
	defer c.resultsLocker.Unlock()

	if _, exists := c.urls[url.Location]; !exists {
		c.urls[url.Location] = url
		c.logger.Debug("Crawler: collected URL", utils.InJSON(url))
		return true
	} else {
		c.logger.Debug("Crawler: URL already collected, skip it", utils.InJSON(url))
		return false
	}
}
