# Makefile
PROJECT := ssg
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
BINARY := ssg
DIST := dist
RELEASES := releases
OUT ?= ..
INCLUDE_GIT ?= 1
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: all clean dist release baseline

all:
	go build -mod=vendor $(LDFLAGS) -o $(BINARY)

clean:
	rm -f $(BINARY)
	rm -rf $(DIST)

dist: clean
	@mkdir -p $(DIST)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		go build -mod=vendor $(LDFLAGS) \
			-o $(DIST)/$(BINARY)-$(VERSION)-$${platform%/*}-$${platform#*/} || exit 1; \
	done
	@echo "Built binaries in $(DIST)/"

release: dist
	@mkdir -p $(RELEASES)
	@cp $(DIST)/* $(RELEASES)/
	@echo "Copied to $(RELEASES)/"

# Snapshot the working tree, including uncommitted source and optionally Git.
# Generated site output, binaries, releases and source maps are excluded.
baseline:
	@mkdir -p '$(OUT)'
	@set -eu; \
	missing=$$(git ls-files --deleted 2>/dev/null || true); \
	if [ -n "$$missing" ]; then \
		echo 'baseline: tracked files missing:'; \
		printf '%s\n' "$$missing"; \
		exit 1; \
	fi; \
	name='$(PROJECT)-baseline-$(VERSION)'; \
	out=$$(cd '$(OUT)' && pwd)/$$name.tar.gz; \
	git_exclude=''; \
	if [ '$(INCLUDE_GIT)' = 0 ]; then \
		git_exclude='--exclude=$(PROJECT)/.git'; \
	fi; \
	cd ..; \
	tar --sort=name --mtime='@0' --owner=0 --group=0 --numeric-owner \
		--exclude='$(PROJECT)/dist' \
		--exclude='$(PROJECT)/docs' \
		--exclude='$(PROJECT)/ssg' \
		--exclude='$(PROJECT)/releases' \
		--exclude='$(PROJECT)/site-src/assets/js/*.map' \
		$$git_exclude -cf - '$(PROJECT)' | gzip -n > "$$out"; \
	echo "baseline: $$out"
