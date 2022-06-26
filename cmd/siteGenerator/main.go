package main

import (
	"fmt"
	"os"
	"sitemap-generator/cmd/siteGenerator/options"
	"sitemap-generator/pkg/crawlers"
	"sitemap-generator/pkg/crawlers/models"
	"sitemap-generator/pkg/parsers"
	"sitemap-generator/pkg/readers"
	"sitemap-generator/pkg/workerPools"
	"sitemap-generator/pkg/writers"
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
	sw := writers.NewSitemapWriter(file)
	crawler := crawlers.NewCrawler(crawlers.CrawlerOptions{
		MaxDepth:      opts.MaxDepth,
		Logger:        logger,
		WorkerPool:    wPool,
		Reader:        reader,
		Parser:        parser,
		SitemapWriter: sw,
	})

	// traverse the start URL recursively
	var urls []*models.Url
	if urls, err = crawler.Traverse(opts.StartUrl); err != nil {
		logger.Fatal("Error while scanning", err.Error())
	}

	// build sitemap XML and write it to the output file
	if err = crawler.WriteSitemap(urls); err != nil {
		logger.Fatal("Error while write to sitemap", err.Error())
	}
}
