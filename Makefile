BINARY  := uqpay
MODULE  := github.com/uqpay/uqpay-cli
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
DATE    := $(shell date +%Y-%m-%d)
LDFLAGS := -s -w -X $(MODULE)/internal/build.Version=$(VERSION) -X $(MODULE)/internal/build.Date=$(DATE)

PREFIX  ?= /usr/local

.PHONY: build test install uninstall clean

build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) .

test:
	CGO_ENABLED=0 go test ./...

install: build
	install -d $(PREFIX)/bin
	install -m755 $(BINARY) $(PREFIX)/bin/$(BINARY)
	@echo "OK: $(PREFIX)/bin/$(BINARY) ($(VERSION))"

uninstall:
	rm -f $(PREFIX)/bin/$(BINARY)

clean:
	rm -f $(BINARY)
