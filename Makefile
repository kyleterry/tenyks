.PHONY: build doc fmt lint run test vendor-get

GO ?= go
PREFIX?=/usr/local
INSTALL_BIN=$(PREFIX)/bin/
TEST_TARGETS=./irc ./service ./config ./control ./mockirc ./cmd/tenyksctl ./version ./

default: all

all: test build

build: tenyks tenyksctl

force-build: clean build

tenyks:
	$(GO) build -v

tenyksctl:
	$(GO) build -v ./cmd/tenyksctl

doc:
	godoc -http=:6060 -index

fmt:
	$(GO) fmt ./src/...

lint:
	golint ./src

run:
	$(GO) run --race tenyks.go

test:
	go test $(TEST_TARGETS)

clean:
	rm -rf tenyks tenyksctl

install: 
	install tenyks $(INSTALL_BIN)tenyks
	install tenyksctl $(INSTALL_BIN)tenyksctl

uninstall:
	rm -rf $(INSTALL_BIN)tenyks

.PHONY: all build force-build doc fmt lint run test clean install uninstall
