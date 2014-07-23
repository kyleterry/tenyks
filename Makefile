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
	$(GO) build -v -o ./bin/tenyks ./tenyks

doc:
	godoc -http=:6060 -index

fmt:
	$(GO) fmt ./src/...

lint:
	golint ./src

run: build
	./bin/tenyks

test:
	$(GO) test ./src/...

clean:
	rm -rf $(BUILDDIR)
	rm -rf ./bin

install: 
	@echo "$(OK_COLOR)===> Installing$(NO_COLOR)"
	install ./bin/tenyks $(INSTALL_BIN)tenyks

uninstall:
	rm -rf $(INSTALL_BIN)tenyks

vendor-get:
	@echo "$(OK_COLOR)===> Fetching dependencies$(NO_COLOR)"
	GOPATH=$(GOPATH) $(GO) get -d -u -v \
	github.com/garyburd/redigo/redis \
	github.com/op/go-logging
