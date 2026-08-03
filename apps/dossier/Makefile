# Dossier — Windows desktop app (built from WSL/Linux)
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

# Production GUI binary (no console window) + console twin.
# Always rebuild both so Dossier-console.exe never lags a tier behind Dossier.exe.
# OriginalFilename in PE Details must match each output name: regenerate winres
# between the two links (shared winres.json cannot hold two OriginalFilename values).
build:
	mkdir -p $(OUT_DIR)
	python3 -c "import json; p='winres/winres.json'; d=json.load(open(p)); d['RT_VERSION']['#1']['0000']['info']['0409']['OriginalFilename']='Dossier.exe'; json.dump(d, open(p,'w'), indent=2); open(p,'a').write('\n')"
	$(WINRES) make --in winres/winres.json --out rsrc --arch amd64
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o $(EXE) .
	python3 -c "import json; p='winres/winres.json'; d=json.load(open(p)); d['RT_VERSION']['#1']['0000']['info']['0409']['OriginalFilename']='Dossier-console.exe'; json.dump(d, open(p,'w'), indent=2); open(p,'a').write('\n')"
	$(WINRES) make --in winres/winres.json --out rsrc --arch amd64
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags="-s -w" -o $(OUT_DIR)/$(APP)-console.exe .
	python3 -c "import json; p='winres/winres.json'; d=json.load(open(p)); d['RT_VERSION']['#1']['0000']['info']['0409']['OriginalFilename']='Dossier.exe'; json.dump(d, open(p,'w'), indent=2); open(p,'a').write('\n')"
	@ls -lh $(EXE) $(OUT_DIR)/$(APP)-console.exe
	@file $(EXE) || true

# Console-attached binary for debugging (see log output)
build-console:
	mkdir -p $(OUT_DIR)
	python3 -c "import json; p='winres/winres.json'; d=json.load(open(p)); d['RT_VERSION']['#1']['0000']['info']['0409']['OriginalFilename']='Dossier-console.exe'; json.dump(d, open(p,'w'), indent=2); open(p,'a').write('\n')"
	$(WINRES) make --in winres/winres.json --out rsrc --arch amd64
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags="-s -w" -o $(OUT_DIR)/$(APP)-console.exe .
	python3 -c "import json; p='winres/winres.json'; d=json.load(open(p)); d['RT_VERSION']['#1']['0000']['info']['0409']['OriginalFilename']='Dossier.exe'; json.dump(d, open(p,'w'), indent=2); open(p,'a').write('\n')"
	@ls -lh $(OUT_DIR)/$(APP)-console.exe

# Headless package verification (Linux). Does not launch WebView2.
run-verify: test
	@echo "Package tests OK. Windows .exe must be launched on Windows host."

# Ship folder: both exes + icon + docs + STATUS if present
release: build
	@mkdir -p releases/$(VERSION)
	cp -f $(EXE) releases/$(VERSION)/Dossier.exe
	cp -f $(OUT_DIR)/$(APP)-console.exe releases/$(VERSION)/Dossier-console.exe
	cp -f assets/dossier-icon.png releases/$(VERSION)/dossier-icon.png
	cp -f VERSION README.md DEV_GUIDE.md USER_GUIDE.md PROJECT_SUMMARY.md releases/$(VERSION)/
	@if [ -f STATUS.md ]; then cp -f STATUS.md releases/$(VERSION)/STATUS.md; fi
	cp -f sample-files/README.txt releases/$(VERSION)/SAMPLE-FILES.txt 2>/dev/null || true
	rm -rf releases/$(VERSION)/sample-files
	cp -a sample-files releases/$(VERSION)/sample-files
	@echo "Release package: releases/$(VERSION)/"
	@ls -la releases/$(VERSION)/

clean:
	rm -f $(SYSO) rsrc_windows_*.syso
	rm -rf $(OUT_DIR)/Dossier.exe $(OUT_DIR)/Dossier-console.exe
