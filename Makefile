.PHONY: build

VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X 'gohandlers/cmd/gohandlers/commands/version.Version=$(VERSION)'"

build:
	@echo "Version $(VERSION)..."
	mkdir -p "build/$(VERSION)"
	GOOS=darwin  GOARCH=amd64 go build -trimpath $(LDFLAGS) -o build/$(VERSION)/gohandlers-$(VERSION)-darwin-amd64  ./cmd/gohandlers
	GOOS=darwin  GOARCH=arm64 go build -trimpath $(LDFLAGS) -o build/$(VERSION)/gohandlers-$(VERSION)-darwin-arm64  ./cmd/gohandlers
	GOOS=linux   GOARCH=amd64 go build -trimpath $(LDFLAGS) -o build/$(VERSION)/gohandlers-$(VERSION)-linux-amd64   ./cmd/gohandlers
	GOOS=linux   GOARCH=386   go build -trimpath $(LDFLAGS) -o build/$(VERSION)/gohandlers-$(VERSION)-linux-386     ./cmd/gohandlers
	GOOS=linux   GOARCH=arm   go build -trimpath $(LDFLAGS) -o build/$(VERSION)/gohandlers-$(VERSION)-linux-arm     ./cmd/gohandlers
	GOOS=linux   GOARCH=arm64 go build -trimpath $(LDFLAGS) -o build/$(VERSION)/gohandlers-$(VERSION)-linux-arm64   ./cmd/gohandlers
	GOOS=freebsd GOARCH=amd64 go build -trimpath $(LDFLAGS) -o build/$(VERSION)/gohandlers-$(VERSION)-freebsd-amd64 ./cmd/gohandlers
	GOOS=freebsd GOARCH=386   go build -trimpath $(LDFLAGS) -o build/$(VERSION)/gohandlers-$(VERSION)-freebsd-386   ./cmd/gohandlers
	GOOS=freebsd GOARCH=arm   go build -trimpath $(LDFLAGS) -o build/$(VERSION)/gohandlers-$(VERSION)-freebsd-arm   ./cmd/gohandlers

.PHONY: install

install:
	go build $(LDFLAGS) -o ~/bin/gohandlers ./cmd/gohandlers

README.toc.md: README.md
	pandoc -s --toc --toc-depth=6 --wrap=preserve README.md -o README.toc.md
	gsed --in-place 's/{.*}//g' README.toc.md
