GOOS ?= linux
ARCH ?= amd64
BUILD := $(shell git describe --always --dirty)
VERSION ?= ${BUILD}

.PHONY: all
all: test build

.PHONY: test
test:
	@go test ./pkg/oneandone

.PHONY: build
build:
	@GOOS=${GOOS} GOARCH=${ARCH} CGO_ENABLED=0 go build \
	-ldflags "-X main.version=${VERSION} -X main.build=${BUILD}" \
	.