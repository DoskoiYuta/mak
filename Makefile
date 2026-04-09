BINARY  := mak
PKG     := github.com/doskoiyuta/mak
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X $(PKG)/cmd.Version=$(VERSION)

GO        ?= go
GOLANGCI  ?= golangci-lint

.PHONY: all build fmt vet lint test tidy clean install run help

## デフォルトターゲット: fmt, vet, lint, test, build
all: fmt vet lint test build

## アプリをビルドする
build:
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BINARY) .

## gofmt を実行する
fmt:
	$(GO) fmt ./...

## go vet を実行する
vet:
	$(GO) vet ./...

## golangci-lint を実行する
lint:
	$(GOLANGCI) run ./...

## go test を実行する
test:
	$(GO) test ./... -race -count=1

## go mod tidy を実行する
tidy:
	$(GO) mod tidy

## ビルド成果物を削除する
clean:
	rm -f $(BINARY)

## $GOPATH/bin へインストールする
install:
	$(GO) install -ldflags "$(LDFLAGS)" .

## mak を起動する
run:
	$(GO) run .

## ヘルプを表示する
help:
	@echo "使用可能なターゲット:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed -e 's/## //'
