GOOS := linux
GOARCH := amd64
CGO_ENABLED := 0

BUILD_DIR := build
BINARY := $(BUILD_DIR)/pizza-son

# Asset cache-busting version. Falls back to "dev" outside a git checkout.
VERSION := $(shell git describe --always --dirty)
ifeq ($(strip $(VERSION)),)
VERSION := dev
endif
LDFLAGS := -s -w -X pizza-son/internal/webui.Version=$(VERSION)

ifneq ($(OS),Windows_NT)
BUILD_ENV = CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH)
CLEAN = rm -rf $(BUILD_DIR)
MKDIR_CSS = mkdir -p internal/webui/public/css
else
BUILD_ENV = set CGO_ENABLED=$(CGO_ENABLED)&& set GOOS=$(GOOS)&& set GOARCH=$(GOARCH)&&
CLEAN = if exist $(BUILD_DIR) rmdir /S /Q $(BUILD_DIR)
MKDIR_CSS = if not exist internal\webui\public\css mkdir internal\webui\public\css
endif

SASS := sass
SASS_FLAGS := --style=compressed --sourcemap=none --no-cache
CSS_SRC := internal/webui/styles/site.scss
CSS_OUT := internal/webui/public/css/site.css

.PHONY: all build css generate clean

all: build

build: css generate
	$(BUILD_ENV) go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/pizza-son/

css:
	$(MKDIR_CSS)
	$(SASS) $(SASS_FLAGS) $(CSS_SRC) $(CSS_OUT)

generate:
	go generate ./internal/webui

clean:
	$(CLEAN)
