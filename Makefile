.PHONY: build doc fmt lint run test vendor-get

GO ?= go
BUILDDIR := $(CURDIR)/build
PREFIX?=/usr/local
INSTALL_BIN=$(PREFIX)/bin/
NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

default: build

build:
	@echo "$(OK_COLOR)===> Building$(NO_COLOR)"
	$(GO) build -o ./bin/tenyks ./tenyks.go
	$(GO) build -o ./bin/tenyksctl ./tenyksctl

doc:
	godoc -http=:6060 -index

fmt:
	$(GO) fmt ./src/...

lint:
	golint ./src

run:
	@echo "$(OK_COLOR)===> Running$(NO_COLOR)"
	$(GO) run --race tenyks.go

clean:
	@echo "$(WARN_COLOR)===> Cleaning$(NO_COLOR)"
	rm -rf $(BUILDDIR)
	rm -rf ./bin

install: 
	@echo "$(OK_COLOR)===> Installing$(NO_COLOR)"
	install ./bin/tenyks $(INSTALL_BIN)tenyks
	install ./bin/tenyksctl $(INSTALL_BIN)tenyksctl

uninstall:
	@echo "$(WARN_COLOR)===> Uninstalling$(NO_COLOR)"
	rm -rf $(INSTALL_BIN)tenyks
