BINARY := link-router

# Windows resource (icon + version metadata) embedded into the .exe.
GOVERSIONINFO := $(shell go env GOPATH)/bin/goversioninfo
SYSO          := resource_windows_amd64.syso
# ImageMagick 7 ships `magick`; ImageMagick 6 ships `convert`. Use whichever exists.
MAGICK        := $(shell command -v magick || command -v convert)

# Version baked into the .exe metadata. Override from CI, e.g. `make build-windows VERSION=1.2.3`.
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
V_PARTS  := $(subst ., ,$(VERSION))
V_MAJOR  := $(word 1,$(V_PARTS))
V_MINOR  := $(word 2,$(V_PARTS))
V_PATCH  := $(word 3,$(V_PARTS))

# Injected into the binary so `link-router version` reports the build version.
# -s -w drop the symbol table and DWARF data; -trimpath keeps local build paths
# out of the shipped binary.
LDFLAGS  := -X main.version=$(VERSION) -s -w
GOFLAGS  := -trimpath

DIST := dist

# Keep in sync with the version pinned in .github/workflows/ci.yml.
GOLANGCI_VERSION := v2.12.2
GOLANGCI         := $(shell go env GOPATH)/bin/golangci-lint

.PHONY: build build-windows dist icon test lint format run install uninstall clean

build:
	CGO_ENABLED=0 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY) .

build-windows: $(SYSO)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY).exe .

# All release artifacts plus a SHA256SUMS manifest, ready to attach to a
# GitHub Release. Used by .github/workflows/release-please.yml.
dist: $(SYSO)
	mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)-linux-arm64 .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)-windows-amd64.exe .
	cd $(DIST) && sha256sum $(BINARY)-* > SHA256SUMS

# Render icon.ico from the source SVG (needs ImageMagick + librsvg).
# icon.ico is committed, so this only re-runs when assets/icon.svg is newer.
icon.ico: assets
	$(MAGICK) -background none assets/icon.svg -define icon:auto-resize=16,32,48,64,128,256 icon.ico

# Convenience alias: `make icon`.
icon: icon.ico

$(SYSO): versioninfo.json icon.ico
	$(GOVERSIONINFO) -64 \
	  -ver-major=$(V_MAJOR) -ver-minor=$(V_MINOR) -ver-patch=$(V_PATCH) \
	  -product-ver-major=$(V_MAJOR) -product-ver-minor=$(V_MINOR) -product-ver-patch=$(V_PATCH) \
	  -file-version="$(VERSION)" -product-version="$(VERSION)" \
	  -o $(SYSO) versioninfo.json

test:
	go test ./...

# Same linter and config CI runs. Installs the pinned binary on first use; run
# `make $(GOLANGCI)` after bumping GOLANGCI_VERSION to pick up a new version.
lint: $(GOLANGCI)
	$(GOLANGCI) run

$(GOLANGCI):
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)

format:
	gofmt -w .

run: build
	./$(BINARY) browsers

install: build
	./$(BINARY) register

uninstall:
	./$(BINARY) unregister

clean:
	rm -f $(BINARY) $(BINARY).exe $(SYSO)
	rm -rf dist/
