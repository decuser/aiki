.PHONY: build clean install run test fmt lint check bless validate smoke smokegold visual aikitest runsamples enginesmoke enginesmokegold enginecoverage rigorous fuzz hooks

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
	./aiki smoke test/behavior/

smokegold: build
	./aiki smoke --gold test/behavior/

visual: build
	./aiki smoke test/visual/

aikitest:
	./aiki test ./...

runsamples: build
	@set -e; \
	for i in extra/samples/*.ai; do \
		echo "RUN $$i"; \
		./aiki "$$i"; \
	done

enginecoverage: build fmt
	./aiki enginesmoke --stage all --coverage test/structure/engine

enginesmoke: build fmt
	./aiki enginesmoke --stage all --check test/structure/engine

enginesmokegold: build fmt
	./aiki enginesmoke --stage all --gold test/structure/engine

# Snapshot-independent correctness checks. This target never writes gold files.
check: build fmt lint test aikitest enginecoverage

# Bless the current, independently checked behavior and engine structure.
# Blessing records a reference state; it is not validation.
bless: check
	./aiki smoke --gold test/behavior/
	./aiki enginesmoke --stage all --gold test/structure/engine

# Full read-only validation against the already-blessed reference state.
# This target must never create or modify gold files.
validate: check
	./aiki smoke test/behavior/
	./aiki enginesmoke --stage all --check test/structure/engine

fuzz:
	go test -fuzz=FuzzLexer ./test/fuzz/ -fuzztime=30s
	go test -fuzz=FuzzParser ./test/fuzz/ -fuzztime=30s

rigorous: validate fuzz

hooks:
	cp hooks/pre-commit .git/hooks/pre-commit
	cp hooks/pre-push .git/hooks/pre-push
	chmod +x .git/hooks/pre-commit .git/hooks/pre-push
	@echo "Git hooks installed"
