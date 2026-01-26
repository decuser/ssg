# Makefile
VERSION := $(shell git describe --tags --always --dirty)
BINARY := ssg
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: all clean

all:
	go build -mod=vendor $(LDFLAGS) -o $(BINARY)

clean:
	rm -f $(BINARY)

