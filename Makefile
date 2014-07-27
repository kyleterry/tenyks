.PHONY: build doc fmt lint run test vendor-get

GO ?= go
BUILDDIR := $(CURDIR)/build
GOPATH := $(BUILDDIR)
export GOPATH
PREFIX?=/usr/local
INSTALL_BIN=$(PREFIX)/bin/
NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

default: build

build-setup:
	@echo "$(OK_COLOR)===> Linking relative packages to GOPATH$(NO_COLOR)"
	@mkdir -p $(GOPATH)/src/github.com/kyleterry
	@test -d "${GOPATH}/src/github.com/kyleterry/tenyks" || ln -s "$(CURDIR)" "$(GOPATH)/src/github.com/kyleterry/tenyks"


build: vendor-get build-setup
	@echo "$(OK_COLOR)===> Building$(NO_COLOR)"
	$(GO) build -o ./bin/tenyks ./tenyks
	@echo "$(OK_COLOR)===> Done building$(NO_COLOR)"

doc:
	godoc -http=:6060 -index

fmt:
	$(GO) fmt ./src/...

lint:
	golint ./src

run:
	@echo "$(OK_COLOR)===> Running$(NO_COLOR)"
	$(GO) run --race tenyks/tenyks.go

clean:
	@echo "$(WARN_COLOR)===> Cleaning$(NO_COLOR)"
	rm -rf $(BUILDDIR)
	rm -rf ./bin

install: 
	@echo "$(OK_COLOR)===> Installing$(NO_COLOR)"
	install ./bin/tenyks $(INSTALL_BIN)tenyks

uninstall:
	@echo "$(WARN_COLOR)===> Uninstalling$(NO_COLOR)"
	rm -rf $(INSTALL_BIN)tenyks

vendor-get:
	@echo "$(OK_COLOR)===> Fetching dependencies$(NO_COLOR)"
	GOPATH=$(GOPATH) $(GO) get -d -u -v \
	github.com/garyburd/redigo/redis \
	github.com/op/go-logging \
	github.com/pebbe/zmq4 \
	code.google.com/p/gomock/gomock

