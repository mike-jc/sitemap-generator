# Sitemap Generator

Simple sitemap (https://www.sitemaps.org) generator as command line tool.

It generates sitemap for the basic URL with a specified max depth and using parallel workers:
* accepts start URL as argument
* recursively navigates by site pages in parallel
* extracts page URLs only from `<a>` elements and take in account `<base>` element if
declared

## CLI flags

Command flags are:
* -h show help of usage
* -v shows current version of the app
* -log-level=`name` level of issuing log messages (error, warn, info, debug)
* -timeout=`duration` allowable timeout for each URL reading (valid duration units are 'ms', 's', 'm')
* -max-retries=`num` max retries for each URL reading
* -max-redirects=`num` max redirects when server response with redirect HTTP response
* -parallel=`num` number of parallel workers to navigate through site
* -max-depth=`num` max depth of url navigation recursion
* -output-file=`path-to-file` output file path

## How to use

1. Download from the repository: 

```shell
    git clone git@github.com:mike-jc/service-scim.git
```

or download a ZIP archive from https://github.com/mike-jc/service-scim

2. Build an executable file

```shell
    cd service-scim
    go mod tidy
    sh build.sh
```

3. Run a command

The mandatory command argument is URL from which the site is started to be scanned.

```shell
    ./build/siteGenerator -parallel=10 --output-file=sitemap.xml https://www.sitemaps.org/
```

## How to test

```shell
    cd service-scim
    go test -timeout 3s ./...
```

## Improvements to be considered

* When halt the app (e.g. by Ctrl+C), write all already found links to sitemap file. Currently it's being written at the end of the traversing

## Technical TODOs

* Use a docker to run a command (Dockerfile, docker-compose.yml)
* Use third-party libraries:
  * For CLI app in Go: Cobra (https://github.com/spf13/cobra), Viper (https://github.com/spf13/viper).
  * For logging: Zap (https://github.com/uber-go/zap)
* Better coverage of unit tests (not only happy flow)