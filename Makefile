VERSION := $(shell git describe --tags --always)
GIT_COMMIT := $(shell git rev-parse --short HEAD)

GO_BUILD_COMMAND := go build -a -ldflags '-s -w -extldflags "-static" -X "github.com/PWZER/govm/cmd.Version=$(VERSION)" -X "github.com/PWZER/govm/cmd.GitCommit=$(GIT_COMMIT)"'

all: darwin linux

compress:
	upx --best bin/* || true

darwin-amd64:
	GOOS=darwin GOARCH=amd64 $(GO_BUILD_COMMAND) -o bin/govm-darwin-amd64 .

darwin-arm64:
	GOOS=darwin GOARCH=arm64 $(GO_BUILD_COMMAND) -o bin/govm-darwin-arm64 .

darwin: darwin-amd64 darwin-arm64

linux-amd64:
	GOOS=linux GOARCH=amd64 $(GO_BUILD_COMMAND) -o bin/govm-linux-amd64 .

linux-arm64:
	GOOS=linux GOARCH=arm64 $(GO_BUILD_COMMAND) -o bin/govm-linux-arm64 .

linux: linux-amd64 linux-arm64
