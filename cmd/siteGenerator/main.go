package main

import (
	"fmt"
	"os"
	"sitemap-generator/cmd/siteGenerator/options"
	"sitemap-generator/pkg/crawlers"
	crawlersModels "sitemap-generator/pkg/crawlers/models"
	"sitemap-generator/pkg/parsers"
	"sitemap-generator/pkg/readers"
	"sitemap-generator/pkg/workerPools"
	"sitemap-generator/pkg/writers"
	writersModels "sitemap-generator/pkg/writers/models"
	"sitemap-generator/services"
	"sitemap-generator/utils"
)

const cmdName = "siteGenerator"

// Version is set during build via --ldflags parameter
var Version = "untagged build"

func main() {
	opts := options.Options{}
	options.ParseOptions(&opts)

	logger, err := services.NewLogger(os.Stderr, cmdName, opts.LogLevel)
	if err != nil {
		logger.Fatal("Can not initialize logger", err.Error())
	}

	options.Validate(logger, opts)
	logger.Info("Started with options", utils.InJSON(opts))

	if opts.ShowVersion {
		fmt.Printf("\nversion is %s\n\n", Version)
		os.Exit(0)
	}

	// check and open output file
	if opts.StartUrl == "" {
		logger.Fatal("Start URL missed. Should be a command argument: siteGenerator <start-url>")
	}
	file, err := os.OpenFile(opts.OutputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		logger.Fatal("Can not open output file", err.Error())
	}

	// create services
	wPool := workerPools.NewWorkerPool(logger, opts.ParallelRoutines)
	reader := readers.NewReader(readers.ReaderOptions{
		Timeout:      opts.Timeout,
		MaxRetries:   opts.MaxRetries,
		MaxRedirects: opts.MaxRedirects,
	})
	parser := parsers.NewParser()
	crawler := crawlers.NewCrawler(crawlers.CrawlerOptions{
		MaxDepth:   opts.MaxDepth,
		Logger:     logger,
		WorkerPool: wPool,
		Reader:     reader,
		Parser:     parser,
	})

	// traverse the start URL recursively
	var urls []*crawlersModels.Url
	if urls, err = crawler.Traverse(opts.StartUrl); err != nil {
		logger.Fatal("Error while scanning", err.Error())
	}

	// build sitemap XML and write it to the output file
	sitemap := writersModels.Sitemap{}
	sitemap.Urls = make([]writersModels.SiteUrl, len(urls))
	for i, u := range urls {
		sitemap.Urls[i] = writersModels.BuildSitemapUrl(u.Location, u.LastModified)
	}

	sw := writers.NewSitemapWriter(file)
	if err = sw.Write(sitemap); err != nil {
		logger.Fatal("Error while write to sitemap", err.Error())
	}
}
