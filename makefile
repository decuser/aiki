.PHONY: build clean install run test fmt doclint lint validate smoke

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X aiki/version.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o aiki ./cmd/aiki

clean:
	rm -f aiki

install: build
	cp aiki ~/bin/

run: build
	./aiki

test:
	go test ./...

fmt:
	go fmt ./...
	./aiki fmt ./...

doclint:
	./aiki doclint

lint:
	./aiki lint ./...

smoke:
	./aiki smoke ./...

validate: build fmt lint doclint test smoke

