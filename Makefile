.PHONY: build doc fmt lint run test vendor-get

GO ?= go
BUILDDIR := $(CURDIR)/build
GOPATH := $(BUILDDIR)
export GOPATH
PREFIX?=/usr/local
INSTALL_BIN=$(PREFIX)/bin/

default: build

build-setup:
	@echo "Linking relative packages to GOPATH"
	@mkdir -p $(GOPATH)/src/github.com/kyleterry
	@test -d "${GOPATH}/src/github.com/kyleterry/tenyks" || ln -s "$(CURDIR)" "$(GOPATH)/src/github.com/kyleterry/tenyks"


build: vendor-get build-setup
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
	install ./bin/tenyks $(INSTALL_BIN)tenyks

uninstall:
	rm -rf $(INSTALL_BIN)tenyks

vendor-get:
	GOPATH=$(GOPATH) $(GO) get -d -u -v \
	github.com/garyburd/redigo/redis \
	github.com/op/go-logging
