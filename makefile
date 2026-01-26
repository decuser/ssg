# Makefile
VERSION := $(shell git describe --tags --always --dirty)
BINARY := ssg
DIST := dist
RELEASES := releases
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: all clean dist release

all:
	go build -mod=vendor $(LDFLAGS) -o $(BINARY)

clean:
	rm -f $(BINARY)
	rm -rf $(DIST)

dist: clean
	@mkdir -p $(DIST)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		go build -mod=vendor $(LDFLAGS) -o $(DIST)/$(BINARY)-$(VERSION)-$${platform%/*}-$${platform#*/}; \
	done
	@echo "Built binaries in $(DIST)/"

release: dist
	@mkdir -p $(RELEASES)
	@cp $(DIST)/* $(RELEASES)/
	@echo "Copied to $(RELEASES)/"
