.PHONY: build clean install run test fmt lint

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X aiki/version.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o aiki ./cmd

clean:
	rm -f aiki

install: build
	cp aiki ~/bin/

run: build
	./aiki

test:
	go test ./tests

fmt:
	go fmt ./...
	./aiki fmt ./...

lint:
	./aiki lint ./...
