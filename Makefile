GOOS := linux
GOARCH := amd64
CGO_ENABLED := 0

BUILD_DIR := build
BINARY := $(BUILD_DIR)/pizza-son

ifeq ($(OS),Windows_NT)
BUILD_ENV = set CGO_ENABLED=$(CGO_ENABLED)&& set GOOS=$(GOOS)&& set GOARCH=$(GOARCH)&&
CLEAN = if exist $(BUILD_DIR) rmdir /S /Q $(BUILD_DIR)
else
BUILD_ENV = CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH)
CLEAN = rm -rf $(BUILD_DIR)
endif

.PHONY: all build generate clean

all: build

build: generate
	$(BUILD_ENV) go build -ldflags "-s -w" -o $(BINARY) ./cmd/pizza-son/

generate:
	go generate ./internal/webui

clean:
	$(CLEAN)
