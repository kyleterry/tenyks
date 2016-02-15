.PHONY: build doc fmt lint run test vendor-get

GO ?= go
PREFIX?=/usr/local
INSTALL_BIN=$(PREFIX)/bin/
NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m
TEST_TARGETS=./irc ./service ./config ./control ./mockirc ./tenyksctl ./version ./

default: all

all: test build

build: ./bin/tenyks ./bin/tenyksctl

./bin/tenyks:
	$(GO) build -v -o ./bin/tenyks ./tenyks.go

./bin/tenyksctl:
	$(GO) build -v -o ./bin/tenyksctl ./tenyksctl

doc:
	godoc -http=:6060 -index

fmt:
	$(GO) fmt ./src/...

lint:
	golint ./src

run:
	@echo "$(OK_COLOR)===> Running$(NO_COLOR)"
	$(GO) run --race tenyks.go

test:
	go test $(TEST_TARGETS)

clean:
	@echo "$(WARN_COLOR)===> Cleaning$(NO_COLOR)"
	rm -rf ./bin

install: 
	@echo "$(OK_COLOR)===> Installing$(NO_COLOR)"
	install ./bin/tenyks $(INSTALL_BIN)tenyks
	install ./bin/tenyksctl $(INSTALL_BIN)tenyksctl

uninstall:
	@echo "$(WARN_COLOR)===> Uninstalling$(NO_COLOR)"
	rm -rf $(INSTALL_BIN)tenyks

.PHONY: all build doc fmt lint run test clean install uninstall
