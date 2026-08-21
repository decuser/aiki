.PHONY: build clean dist dist-linux-amd64 dist-windows-amd64 distcheck baseline run test fmt lint treecheck historycheck check bless validate smoke smokegold visual aikitest runsamples enginesmoke enginesmokegold enginecoverage conformance selfhost rigorous fuzz hooks profilesweep install-xed-plugin uninstall-xed-plugin install-vscode-plugin uninstall-vscode-plugin

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

DIST_PARENT ?= $(abspath ..)

LINUX_DIST_NAME := aiki-$(VERSION)-linux-amd64
LINUX_DIST_DIR := $(DIST_PARENT)/$(LINUX_DIST_NAME)
LINUX_DIST_ARCHIVE := $(DIST_PARENT)/$(LINUX_DIST_NAME).tar.gz

WINDOWS_DIST_NAME := aiki-$(VERSION)-windows-amd64
WINDOWS_DIST_DIR := $(DIST_PARENT)/$(WINDOWS_DIST_NAME)
WINDOWS_DIST_ARCHIVE := $(DIST_PARENT)/$(WINDOWS_DIST_NAME).tar.gz

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
	go build -o aiki-canvas ./cmd/aiki-canvas

clean:
	rm -f aiki aiki-canvas
	rm -rf "$(LINUX_DIST_DIR)" "$(WINDOWS_DIST_DIR)"
	rm -f "$(LINUX_DIST_ARCHIVE)" "$(WINDOWS_DIST_ARCHIVE)"
	rm -f "$(BASELINE_ARCHIVE)"

# Build all supported Alpha distributions beside the source tree.
# Linux and Windows receive the same library/experiment payload; only the
# executable differs.
dist: dist-linux-amd64 dist-windows-amd64

# Build the supported Linux amd64 distribution. The executable lives at the
# distribution root so users can add that one directory to PATH.
dist-linux-amd64:
	@set -eu; \
	rm -rf "$(LINUX_DIST_DIR)"; \
	rm -f "$(LINUX_DIST_ARCHIVE)"; \
	mkdir -p "$(LINUX_DIST_DIR)"; \
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o "$(LINUX_DIST_DIR)/aiki" ./cmd/aiki; \
	GOOS=linux GOARCH=amd64 go build -o "$(LINUX_DIST_DIR)/aiki-canvas" ./cmd/aiki-canvas; \
	cp LICENSE README.md "$(LINUX_DIST_DIR)/"; \
	cp -R lib "$(LINUX_DIST_DIR)/lib"; \
	cp -R experiments "$(LINUX_DIST_DIR)/experiments"; \
	if [ -d vendor ]; then cp -R vendor "$(LINUX_DIST_DIR)/vendor"; fi; \
	tar -C "$(DIST_PARENT)" -czf "$(LINUX_DIST_ARCHIVE)" "$(LINUX_DIST_NAME)"; \
	echo "$(LINUX_DIST_DIR)"; \
	echo "$(LINUX_DIST_ARCHIVE)"

# Build the supported Windows amd64 distribution. Keep tar.gz for both targets
# so release packaging has no additional zip-tool dependency.
dist-windows-amd64:
	@set -eu; \
	rm -rf "$(WINDOWS_DIST_DIR)"; \
	rm -f "$(WINDOWS_DIST_ARCHIVE)"; \
	mkdir -p "$(WINDOWS_DIST_DIR)"; \
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o "$(WINDOWS_DIST_DIR)/aiki.exe" ./cmd/aiki; \
	GOOS=windows GOARCH=amd64 go build -o "$(WINDOWS_DIST_DIR)/aiki-canvas.exe" ./cmd/aiki-canvas; \
	cp LICENSE README.md "$(WINDOWS_DIST_DIR)/"; \
	cp -R lib "$(WINDOWS_DIST_DIR)/lib"; \
	cp -R experiments "$(WINDOWS_DIST_DIR)/experiments"; \
	if [ -d vendor ]; then cp -R vendor "$(WINDOWS_DIST_DIR)/vendor"; fi; \
	tar -C "$(DIST_PARENT)" -czf "$(WINDOWS_DIST_ARCHIVE)" "$(WINDOWS_DIST_NAME)"; \
	echo "$(WINDOWS_DIST_DIR)"; \
	echo "$(WINDOWS_DIST_ARCHIVE)"

# Prove the published Linux archive is independent of both the source tree and
# the sibling staging directory. Building dist also proves the supported
# Windows package cross-compiles and is structurally complete.
distcheck: dist
	@set -eu; \
	tmp=$$(mktemp -d); \
	trap 'rm -rf "$$tmp"' EXIT; \
	tar -xzf "$(LINUX_DIST_ARCHIVE)" -C "$$tmp"; \
	test -f "$$tmp/$(LINUX_DIST_NAME)/experiments/README.md" || { echo "distcheck: experiments collection missing from Linux archive" >&2; exit 1; }; \
	mkdir -p "$$tmp/work/one" "$$tmp/work/two"; \
	printf '%s\n' 'package decoy' > "$$tmp/work/one/decoy.ai"; \
	printf '%s\n' 'package decoy' > "$$tmp/work/two/decoy.ai"; \
	printf '%s\n' 'use("list")' 'println(sum([1, 2, 3]))' > "$$tmp/work/check.ai"; \
	cd "$$tmp/work"; \
	PATH="$$tmp/$(LINUX_DIST_NAME):$$PATH" aiki check.ai > output.txt; \
	grep -qx '6' output.txt; \
	created=$$(PATH="$$tmp/$(LINUX_DIST_NAME):$$PATH" aiki experiment new "Distribution probe" | sed -n 's/^created //p'); \
	test -n "$$created"; \
	test -f "$$created/README.md"; \
	test -f "$$created/experiment/PROCEDURE.md"; \
	test -x "$$created/experiment/run.sh"; \
	test -d "$$created/results"; \
	test -d "$$created/analyses"; \
	test -x "$$tmp/$(LINUX_DIST_NAME)/aiki-canvas" || { echo "distcheck: Linux archive missing aiki-canvas" >&2; exit 1; }; \
	tar -tzf "$(WINDOWS_DIST_ARCHIVE)" | grep -qx "$(WINDOWS_DIST_NAME)/aiki.exe" || { echo "distcheck: Windows archive missing aiki.exe" >&2; exit 1; }; \
	tar -tzf "$(WINDOWS_DIST_ARCHIVE)" | grep -qx "$(WINDOWS_DIST_NAME)/aiki-canvas.exe" || { echo "distcheck: Windows archive missing aiki-canvas.exe" >&2; exit 1; }; \
	tar -tzf "$(WINDOWS_DIST_ARCHIVE)" | grep -qx "$(WINDOWS_DIST_NAME)/lib/" || { echo "distcheck: Windows archive missing lib" >&2; exit 1; }; \
	echo "distcheck ok: Linux archive is relocatable and Windows archive cross-builds with the shipped library"

# Capture a portable development baseline beside the source tree. Unlike dist,
# this is a repository snapshot: it intentionally includes .git so branch,
# history, refs, and session state travel with the source. Only the built
# top-level release executables are omitted.
baseline:
	@set -eu; \
	rm -f "$(BASELINE_ARCHIVE)"; \
	tar -C "$(DIST_PARENT)" \
		--exclude="$(SOURCE_DIR_NAME)/aiki" \
		--exclude="$(SOURCE_DIR_NAME)/aiki-canvas" \
		--exclude="$(SOURCE_DIR_NAME)/aiki.exe" \
		--exclude="$(SOURCE_DIR_NAME)/aiki-canvas.exe" \
		-czf "$(BASELINE_ARCHIVE)" "$(SOURCE_DIR_NAME)"; \
	tar -tzf "$(BASELINE_ARCHIVE)" | grep -q "^$(SOURCE_DIR_NAME)/\.git/" || { \
		echo "baseline: archive is missing .git" >&2; exit 1; \
	}; \
	if tar -tzf "$(BASELINE_ARCHIVE)" | grep -Eq "^$(SOURCE_DIR_NAME)/(aiki|aiki-canvas)(\.exe)?$$"; then \
		echo "baseline: archive unexpectedly contains built executable" >&2; exit 1; \
	fi; \
	echo "$(BASELINE_ARCHIVE)"

run: build
	./aiki

# Behavioral and implementation tests. Architectural invariants are run
# separately by `make invariant` and composed back into `make check`.
test:
	@set -e; \
	pkgs="$$(go list ./... | grep -Ev '/test/(invariant|boundary|conformance)$$')"; \
	go test $$pkgs

# Fast architectural contract checks. Keep this target free of fuzzing, gold
# blessing, and other long-running behavioral validation.
invariant:
	go test ./test/invariant ./test/boundary

fmt:
	go fmt ./...
	./aiki fmt ./...
	./aiki distfmt ./...

lint:
	./aiki lint ./...

treecheck: build
	./aiki treecheck

historycheck:
	@hooks/check-history-cruft

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
check: build fmt lint treecheck test invariant aikitest enginecoverage

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

conformance: build
	go test ./test/conformance

selfhost: build
	go test -tags rigorous ./test/conformance -run Selfhost

fuzz:
	go test -fuzz=FuzzLexer ./test/fuzz/ -fuzztime=30s
	go test -fuzz=FuzzParser ./test/fuzz/ -fuzztime=30s

rigorous: validate conformance selfhost fuzz

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
