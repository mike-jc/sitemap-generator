package options

import (
	"flag"
	"sitemap-generator/services"
	"time"
)

const (
	logLevel        = "log-level"
	logLevelDefault = "info"

	timeout        = "timeout"
	timeoutDefault = 5 * time.Second

	maxRetries        = "max-retries"
	maxRetriesDefault = 3

	maxRedirects        = "max-redirects"
	maxRedirectsDefault = 10

	parallel        = "parallel"
	parallelDefault = 5

	maxDepth        = "max-depth"
	maxDepthDefault = 2

	outputFile        = "output-file"
	outputFileDefault = "sitemap.xml"
)

type Options struct {
	ShowVersion      bool          `json:"showVersion"`
	LogLevel         string        `json:"logLevel"`
	Timeout          time.Duration `json:"timeout"`
	MaxRetries       int           `json:"maxRetries"`
	MaxRedirects     int           `json:"maxRedirects"`
	ParallelRoutines int           `json:"parallelRoutines"`
	MaxDepth         int           `json:"maxDepth"`
	OutputFile       string        `json:"outputFile"`
	StartUrl         string        `json:"startUrl"`
}

func ParseOptions(opts *Options) {
	flag.BoolVar(&opts.ShowVersion, "v", false, "show version")
	flag.StringVar(&opts.LogLevel, logLevel, logLevelDefault, "log level (error, warn, info, debug)")
	flag.DurationVar(&opts.Timeout, timeout, timeoutDefault, "allowable timeout for each URL reading (valid duration units are 'ms', 's', 'm')")
	flag.IntVar(&opts.MaxRetries, maxRetries, maxRetriesDefault, "max retries for each URL reading")
	flag.IntVar(&opts.MaxRedirects, maxRedirects, maxRedirectsDefault, "max redirects when server response with redirect HTTP response")
	flag.IntVar(&opts.ParallelRoutines, parallel, parallelDefault, "number of parallel workers to navigate through site")
	flag.IntVar(&opts.MaxDepth, maxDepth, maxDepthDefault, "max depth of URL navigation recursion")
	flag.StringVar(&opts.OutputFile, outputFile, outputFileDefault, "output file path")
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		opts.StartUrl = args[0]
	}

}

func Validate(logger services.Logger, opts Options) {
	if opts.MaxRetries <= 0 {
		logger.Fatal("MaxRetries should be number greater than zero", opts)
	}
}
