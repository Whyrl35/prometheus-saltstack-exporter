# Variables
GO    := GO19VENDOREXPERIMENT=1 GOOS=linux GOARCH=amd64 go
PROMU := $(GOPATH)/bin/$(NAME)
pkgs   = $(shell $(GO) list ./... | grep -v /vendor/)

PREFIX                  ?= $(shell pwd)
BIN_DIR                 ?= $(shell pwd)
#DOCKER_REPO             ?= fxinnovation
#DOCKER_IMAGE_NAME       ?= exporter-template
#DOCKER_IMAGE_TAG        ?= $(subst /,-,$(shell git rev-parse --abbrev-ref HEAD))

# Default target
all: format build test

test: build
	@echo ">> running tests"
	@$(GO) test -short $(pkgs)

style:
	@echo ">> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

format:
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

vet:
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

dependencies:
	rm -rf Gopkg.lock vendor/
	dep ensure

build: promu
	@echo ">> building binaries"
	@$(PROMU) build --prefix $(PREFIX)

tarball: promu
	@echo ">> building release tarball"
	@$(PROMU) tarball --prefix $(PREFIX) $(BIN_DIR)

#docker:
#	@echo ">> building docker image"
#	@docker build -t "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)" .

#dockerlint:
#	@echo ">> linting Dockerfile"
#	@docker run --rm -i hadolint/hadolint < Dockerfile

promu:
	@GOOS=$(shell uname -s | tr A-Z a-z) \
		GOARCH=$(subst x86_64,amd64,$(patsubst i%86,386,$(shell uname -m))) \
		$(GO) get -u github.com/prometheus/promu

run:
	@$(GO) run

clean:
	rm -f bin/*

.PHONY: all style format dependencies build test vet tarball run clean
