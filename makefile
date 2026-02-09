.PHONY: build clean install run test

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o aiki .

clean:
	rm -f aiki

install: build
	cp aiki ~/bin/

run: build
	./aiki

test:
	go test ./tests
