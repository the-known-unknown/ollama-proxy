BINARY := ollama-proxy
PKG := ./cmd/ollama-proxy
BIN_DIR := bin
VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)

PLATFORMS := \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64

.PHONY: all build run test vet fmt clean release

all: build

build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) $(PKG)

run:
	go run $(PKG) $(ARGS)

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

clean:
	rm -rf $(BIN_DIR)

release:
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; arch=$${platform#*/}; \
		out=$(BIN_DIR)/$(BINARY)_$${os}_$${arch}; \
		echo "building $$out"; \
		GOOS=$$os GOARCH=$$arch go build -ldflags "$(LDFLAGS)" -o $$out $(PKG); \
	done
