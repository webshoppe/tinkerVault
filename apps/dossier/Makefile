# Dossier; Windows desktop app (built from WSL/Linux)
GO      ?= $(HOME)/.local/go/bin/go
WINRES  ?= $(HOME)/go/bin/go-winres
export PATH := $(HOME)/.local/go/bin:$(HOME)/go/bin:$(PATH)
export GOPATH ?= $(HOME)/go

APP     := Dossier
OUT_DIR := build
EXE     := $(OUT_DIR)/$(APP).exe
VERSION := $(shell cat VERSION 2>/dev/null || echo 1.0.0)
LDFLAGS := -H windowsgui -s -w
SYSO    := rsrc_windows_amd64.syso

.PHONY: all build build-console test tidy resources release clean

all: test build

tidy:
	$(GO) mod tidy

# Embed icon + version into PE resources (picked up automatically by go build)
resources: $(SYSO)

$(SYSO): winres/winres.json winres/icon_16.png winres/icon_32.png winres/icon_48.png winres/icon_64.png winres/icon_128.png winres/icon_256.png
	$(WINRES) make --in winres/winres.json --out rsrc --arch amd64

test:
	$(GO) test ./internal/... -count=1 -timeout 60s

# Production GUI binary (no console window)
build: resources
	mkdir -p $(OUT_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o $(EXE) .
	@ls -lh $(EXE)
	@file $(EXE) || true

# Console-attached binary for debugging (see log output)
build-console: resources
	mkdir -p $(OUT_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags="-s -w" -o $(OUT_DIR)/$(APP)-console.exe .
	@ls -lh $(OUT_DIR)/$(APP)-console.exe

# Headless package verification (Linux). Does not launch WebView2.
run-verify: test
	@echo "Package tests OK. Windows .exe must be launched on Windows host."

# Ship folder: exe + icon + docs
release: build
	@mkdir -p releases/$(VERSION)
	cp -f $(EXE) releases/$(VERSION)/Dossier.exe
	cp -f assets/dossier-icon.png releases/$(VERSION)/dossier-icon.png
	cp -f VERSION README.md DEV_GUIDE.md USER_GUIDE.md PROJECT_SUMMARY.md releases/$(VERSION)/
	cp -f sample-files/README.txt releases/$(VERSION)/SAMPLE-FILES.txt 2>/dev/null || true
	@# include sample-files for end users
	rm -rf releases/$(VERSION)/sample-files
	cp -a sample-files releases/$(VERSION)/sample-files
	@echo "Release package: releases/$(VERSION)/"
	@ls -la releases/$(VERSION)/

clean:
	rm -f $(SYSO) rsrc_windows_*.syso
	rm -rf $(OUT_DIR)/Dossier.exe $(OUT_DIR)/Dossier-console.exe
