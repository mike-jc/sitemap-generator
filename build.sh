#!/bin/bash
set -e

COMMIT_HASH=$(git rev-parse --short HEAD)
NAME=siteGenerator

mkdir -p ./build
rm -f "$(pwd)/build/${NAME}"

CGO_ENABLED=0 go build -o "./build/${NAME}" -ldflags "-X main.Version=$COMMIT_HASH" ./cmd/${NAME}/
