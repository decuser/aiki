.PHONY: build clean dist distcheck run test fmt lint treecheck check bless validate smoke smokegold visual aikitest runsamples enginesmoke enginesmokegold enginecoverage rigorous fuzz hooks profilesweep

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

DIST_PARENT ?= $(abspath ..)
DIST_OS ?= $(shell go env GOOS)
DIST_ARCH ?= $(shell go env GOARCH)
DIST_NAME := aiki-$(VERSION)-$(DIST_OS)-$(DIST_ARCH)
DIST_DIR := $(DIST_PARENT)/$(DIST_NAME)
DIST_ARCHIVE := $(DIST_PARENT)/$(DIST_NAME).tar.gz

build:
	go build $(LDFLAGS) -o aiki ./cmd/aiki

clean:
	rm -f aiki
	rm -rf "$(DIST_DIR)"
	rm -f "$(DIST_ARCHIVE)"

# Build a relocatable user distribution beside the source tree. The executable
# lives at the distribution root so users can add that one directory to PATH.
dist: build
	@set -eu; \
	rm -rf "$(DIST_DIR)"; \
	rm -f "$(DIST_ARCHIVE)"; \
	mkdir -p "$(DIST_DIR)"; \
	cp aiki LICENSE README.md "$(DIST_DIR)/"; \
	cp -R lib "$(DIST_DIR)/lib"; \
	if [ -d vendor ]; then cp -R vendor "$(DIST_DIR)/vendor"; fi; \
	tar -C "$(DIST_PARENT)" -czf "$(DIST_ARCHIVE)" "$(DIST_NAME)"; \
	echo "$(DIST_DIR)"; \
	echo "$(DIST_ARCHIVE)"

# Prove the published archive is independent of both the source tree and the
# sibling staging directory: unpack it under a temporary prefix, run from an
# unrelated directory via PATH, ignore decoy packages beneath that directory,
# and use a shipped library module.
distcheck: dist
	@set -eu; \
	tmp=$$(mktemp -d); \
	trap 'rm -rf "$$tmp"' EXIT; \
	tar -xzf "$(DIST_ARCHIVE)" -C "$$tmp"; \
	mkdir -p "$$tmp/work/one" "$$tmp/work/two"; \
	printf '%s\n' 'package decoy' > "$$tmp/work/one/decoy.ai"; \
	printf '%s\n' 'package decoy' > "$$tmp/work/two/decoy.ai"; \
	printf '%s\n' 'use("list")' 'println(sum([1, 2, 3]))' > "$$tmp/work/check.ai"; \
	cd "$$tmp/work"; \
	PATH="$$tmp/$(DIST_NAME):$$PATH" aiki check.ai > output.txt; \
	grep -qx '6' output.txt; \
	echo "distcheck ok: relocatable archive loads shipped modules outside source tree"

run: build
	./aiki

test:
	go test ./...

fmt:
	go fmt ./...
	./aiki fmt ./...

lint:
	./aiki lint ./...

treecheck: build
	./aiki treecheck

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
check: build fmt lint treecheck test aikitest enginecoverage

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


PROFILE_DIR ?= profile-out

# Reproducible semantic/substrate profiling sweep. Not part of validation.
profilesweep: build
	@extra/profiling/sweep.sh $(PROFILE_DIR)
