.PHONY: build clean install run test fmt doclint lint validate smoke test-engine

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

doclint:
	-./aiki doclint extra/doc extra/status

lint:
	-./aiki lint ./...

smoke:
	./aiki smoke ./...

validate: build fmt lint doclint test smoke

test-engine:
	go build -o /tmp/aiki-engine cmd/engine-test/main.go
	@for f in tests/smoke/*.ai; do \
		in="$${f%.ai}.in"; \
		echo "=== $$f ==="; \
		if [ -f "$$in" ]; then \
			/tmp/aiki-engine "$$f" < "$$in" 2>&1 | head -5; \
		else \
			/tmp/aiki-engine "$$f" 2>&1 | head -5; \
		fi; \
	done

