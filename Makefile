BINARY = knockr
VERSION = $(shell git describe --tags --abbrev=0)
LDFLAGS = -s -w -X 'main.version=$(VERSION)'

# default install is per-user; use: sudo make install PREFIX=/usr/local
PREFIX ?= $(HOME)/.local
BINDIR = $(PREFIX)/bin
MANDIR = $(PREFIX)/share/man/man1

build:
	go build -v -ldflags="$(LDFLAGS)"

nocgo:
	CGO_ENABLED=0 go build -v -ldflags="$(LDFLAGS)"

install: build man
	mkdir -p $(BINDIR) $(MANDIR)
	install -m 755 $(BINARY) $(BINDIR)/$(BINARY)
	install -m 644 knockr.1 $(MANDIR)/knockr.1

uninstall:
	rm -f $(BINDIR)/$(BINARY) $(MANDIR)/knockr.1

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	go test -race ./...

# full gauntlet: build, vet, lint, race tests
check: build vet lint test

man:
	scdoc < knockr.1.scd > knockr.1

clean:
	rm -f ${BINARY}

.PHONY: build nocgo install uninstall vet lint test check man clean
