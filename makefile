.PHONY: build clean install run test fmt lint validate smoke runsamples enginesmoke

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

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

lint:
	./aiki lint ./...

smoke:
	./aiki smoke ./...

runsamples: build
	@set -e; \
	for i in extra/samples/*.ai; do \
		echo "RUN $$i"; \
		./aiki "$$i"; \
	done

enginesmoke: build fmt
	./aiki enginesmoke --stage all --check tests/engine

enginesmokegold: build fmt
	./aiki enginesmoke --stage all --gold tests/engine

validate: build fmt lint test smoke runsamples enginesmoke

