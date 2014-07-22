.PHONY: build doc fmt lint run test vendor_clean vendor_get vendor_update vet

GOPATH := ${PWD}/_vendor:${GOPATH}
export GOPATH
PREFIX?=/usr/local
INSTALL_BIN=$(PREFIX)/bin/

default: build

build: vet
	go build -v -o ./bin/tenyks ./tenyks.go

doc:
	godoc -http=:6060 -index

fmt:
	go fmt ./src/...

lint:
	golint ./src

run: build
	./bin/tenyks

test:
	go test ./src/...

clean: vendor_clean
	rm -rf ./bin

install: 
	install ./bin/tenyks $(INSTALL_BIN)tenyks

uninstall:
	rm -rf $(INSTALL_BIN)tenyks

vendor_clean:
	rm -dRf ./_vendor/src

vendor_get: vendor_clean
	GOPATH=${PWD}/_vendor go get -d -u -v \
	github.com/op/go-logging

vendor_update: vendor_get
	rm -rf `find ./_vendor/src -type d -name .git` \
	&& rm -rf `find ./_vendor/src -type d -name .hg` \
	&& rm -rf `find ./_vendor/src -type d -name .bzr` \
	&& rm -rf `find ./_vendor/src -type d -name .svn`

vet:
	go vet .
