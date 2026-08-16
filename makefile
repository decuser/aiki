.PHONY: build clean dist distcheck baseline run test fmt lint treecheck check bless validate smoke smokegold visual aikitest runsamples enginesmoke enginesmokegold enginecoverage rigorous fuzz hooks profilesweep install-xed-plugin uninstall-xed-plugin install-vscode-plugin uninstall-vscode-plugin

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

DIST_PARENT ?= $(abspath ..)
DIST_OS ?= $(shell go env GOOS)
DIST_ARCH ?= $(shell go env GOARCH)
DIST_NAME = aiki-$(VERSION)-$(DIST_OS)-$(DIST_ARCH)
DIST_DIR = $(DIST_PARENT)/$(DIST_NAME)
DIST_ARCHIVE = $(DIST_PARENT)/$(DIST_NAME).tar.gz

SOURCE_DIR_NAME := $(notdir $(CURDIR))
BASELINE_NAME := aiki-baseline-$(VERSION)
BASELINE_ARCHIVE := $(DIST_PARENT)/$(BASELINE_NAME).tar.gz

USER_DATA_DIR ?= $(if $(XDG_DATA_HOME),$(XDG_DATA_HOME),$(HOME)/.local/share)
XED_PLUGIN_DIR ?= $(USER_DATA_DIR)/xed/plugins
XED_LANG_DIR ?= $(USER_DATA_DIR)/gtksourceview-4/language-specs

NPM ?= npm
NPX ?= npx
CODE ?= code
VSCODE_AIKI_EXTENSION_ID ?= aiki.aiki-language-services
VSCODE_VSCE ?= @vscode/vsce@3.9.2

build:
	go build $(LDFLAGS) -o aiki ./cmd/aiki

clean:
	rm -f aiki
	rm -rf "$(DIST_DIR)"
	rm -f "$(DIST_ARCHIVE)"
	rm -f "$(BASELINE_ARCHIVE)"

# Build a relocatable user distribution beside the source tree. The executable
# lives at the distribution root so users can add that one directory to PATH.
dist: build
	@set -eu; \
	rm -rf "$(DIST_DIR)"; \
	rm -f "$(DIST_ARCHIVE)"; \
	mkdir -p "$(DIST_DIR)"; \
	cp aiki LICENSE README.md "$(DIST_DIR)/"; \
	cp -R lib "$(DIST_DIR)/lib"; \
	cp -R experiments "$(DIST_DIR)/experiments"; \
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
	test -f "$$tmp/$(DIST_NAME)/experiments/README.md" || { echo "distcheck: experiments collection missing from archive" >&2; exit 1; }; \
	mkdir -p "$$tmp/work/one" "$$tmp/work/two"; \
	printf '%s\n' 'package decoy' > "$$tmp/work/one/decoy.ai"; \
	printf '%s\n' 'package decoy' > "$$tmp/work/two/decoy.ai"; \
	printf '%s\n' 'use("list")' 'println(sum([1, 2, 3]))' > "$$tmp/work/check.ai"; \
	cd "$$tmp/work"; \
	PATH="$$tmp/$(DIST_NAME):$$PATH" aiki check.ai > output.txt; \
	grep -qx '6' output.txt; \
	created=$$(PATH="$$tmp/$(DIST_NAME):$$PATH" aiki experiment new "Distribution probe" | sed -n 's/^created //p'); \
	test -n "$$created"; \
	test -f "$$created/README.md"; \
	test -x "$$created/run.sh"; \
	echo "distcheck ok: relocatable archive loads shipped modules and scaffolds out-of-tree experiments"

# Capture a portable development baseline beside the source tree. Unlike dist,
# this is a repository snapshot: it intentionally includes .git so branch,
# history, refs, and session state travel with the source. Only the built
# top-level aiki executable is omitted.
baseline:
	@set -eu; \
	rm -f "$(BASELINE_ARCHIVE)"; \
	tar -C "$(DIST_PARENT)" \
		--exclude="$(SOURCE_DIR_NAME)/aiki" \
		-czf "$(BASELINE_ARCHIVE)" "$(SOURCE_DIR_NAME)"; \
	tar -tzf "$(BASELINE_ARCHIVE)" | grep -q "^$(SOURCE_DIR_NAME)/\.git/" || { \
		echo "baseline: archive is missing .git" >&2; exit 1; \
	}; \
	if tar -tzf "$(BASELINE_ARCHIVE)" | grep -qx "$(SOURCE_DIR_NAME)/aiki"; then \
		echo "baseline: archive unexpectedly contains built aiki executable" >&2; exit 1; \
	fi; \
	echo "$(BASELINE_ARCHIVE)"

run: build
	./aiki

test:
	go test ./...

fmt:
	go fmt ./...
	./aiki fmt ./...
	./aiki distfmt ./...

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

# Install the user-local Xed integration from the repository. Remove Aiki's
# previous installed copies first so renamed/deleted plugin files cannot linger.
# XED_PLUGIN_DIR and XED_LANG_DIR may be overridden for nonstandard installs.
install-xed-plugin:
	@set -eu; \
	plugin_dir="$(XED_PLUGIN_DIR)"; \
	lang_dir="$(XED_LANG_DIR)"; \
	test -n "$$plugin_dir"; \
	test -n "$$lang_dir"; \
	rm -rf "$$plugin_dir/aiki_lsp"; \
	rm -f "$$plugin_dir/aiki_lsp.plugin" "$$lang_dir/aiki.lang"; \
	mkdir -p "$$plugin_dir/aiki_lsp" "$$lang_dir"; \
	install -m 0644 extra/editors/xed/aiki.lang "$$lang_dir/aiki.lang"; \
	install -m 0644 extra/editors/xed/aiki_lsp.plugin "$$plugin_dir/aiki_lsp.plugin"; \
	install -m 0644 extra/editors/xed/aiki_lsp/__init__.py "$$plugin_dir/aiki_lsp/__init__.py"; \
	echo "Installed Xed Aiki language definition: $$lang_dir/aiki.lang"; \
	echo "Installed Xed Aiki LSP plugin: $$plugin_dir/aiki_lsp.plugin"; \
	echo "Restart Xed and enable 'Aiki Language Services' in Plugins."

uninstall-xed-plugin:
	@set -eu; \
	plugin_dir="$(XED_PLUGIN_DIR)"; \
	lang_dir="$(XED_LANG_DIR)"; \
	test -n "$$plugin_dir"; \
	test -n "$$lang_dir"; \
	rm -rf "$$plugin_dir/aiki_lsp"; \
	rm -f "$$plugin_dir/aiki_lsp.plugin" "$$lang_dir/aiki.lang"; \
	echo "Removed user-local Xed Aiki integration."


PROFILE_DIR ?= profile-out

# Reproducible semantic/substrate profiling sweep. Not part of validation.
profilesweep: build
	@extra/profiling/sweep.sh $(PROFILE_DIR)


# Build and install the thin VS Code client entirely out of tree. npm
# dependencies, package-lock data, and the VSIX live only in a disposable
# staging directory; VS Code itself performs the supported extension install.
install-vscode-plugin:
	@set -eu; \
	command -v "$(NPM)" >/dev/null 2>&1 || { echo "install-vscode-plugin: npm not found" >&2; exit 1; }; \
	command -v "$(NPX)" >/dev/null 2>&1 || { echo "install-vscode-plugin: npx not found" >&2; exit 1; }; \
	command -v "$(CODE)" >/dev/null 2>&1 || { echo "install-vscode-plugin: code not found" >&2; exit 1; }; \
	tmp=$$(mktemp -d); \
	trap 'rm -rf "$$tmp"' EXIT; \
	cp -R extra/editors/vscode/. "$$tmp/"; \
	cp LICENSE "$$tmp/LICENSE"; \
	cd "$$tmp"; \
	"$(NPM)" install --omit=dev --ignore-scripts --no-audit --no-fund; \
	"$(NPX)" --yes "$(VSCODE_VSCE)" package --allow-missing-repository >/dev/null; \
	vsix=$$(find "$$tmp" -maxdepth 1 -name '*.vsix' -print -quit); \
	test -n "$$vsix" || { echo "install-vscode-plugin: VSIX was not produced" >&2; exit 1; }; \
	"$(CODE)" --install-extension "$$vsix" --force; \
	echo "Installed VS Code Aiki extension: $(VSCODE_AIKI_EXTENSION_ID)"; \
	echo "Restart VS Code. Set 'aiki.server.path' if desktop VS Code cannot find the development aiki executable."

uninstall-vscode-plugin:
	@set -eu; \
	command -v "$(CODE)" >/dev/null 2>&1 || { echo "uninstall-vscode-plugin: code not found" >&2; exit 1; }; \
	"$(CODE)" --uninstall-extension "$(VSCODE_AIKI_EXTENSION_ID)"; \
	echo "Removed VS Code Aiki extension: $(VSCODE_AIKI_EXTENSION_ID)"
