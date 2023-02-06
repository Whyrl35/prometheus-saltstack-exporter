GO    := GO19VENDOREXPERIMENT=1 GOOS=linux GOARCH=amd64 go
PROMU := $(GOPATH)/bin/promu
pkgs   = $(shell $(GO) list ./... | grep -v /vendor/)

PREFIX                  ?= $(shell pwd)
BIN_DIR                 ?= $(shell pwd)
VERSION 				:= $(shell cat VERSION)-$(shell git rev-parse --short HEAD)

BINARY_NAME				:= $(shell promu info | grep -oP 'Name: \K.+')
DEB_PACKAGE_NAME		:= $(BINARY_NAME)
DEB_PACKAGE_DESCRIPTION = Prometheus exporter for Saltstack
RPM_PACKAGE_NAME		:= $(BINARY_NAME)
RPM_PACKAGE_DESCRIPTION = Prometheus exporter for Saltstack
BUILD_ARTIFACTS_DIR		= artifacts

all: format style vet build test

test:				# Testing Go code
test: build
	@echo ">> running tests"
	@$(GO) test -short $(pkgs)

style:				# Check style
	@echo ">> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

format:				# Check go fmt
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

vet:				# Run go vet
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

build: 				# Build the binary from go source code
build: promu
	@echo ">> building binaries"
	@$(PROMU) build

## Need ... / fpm tools
build-deb:			# Build DEB package (needs other tools) see: https://echorand.me/posts/building-golang-deb-packages/
	test $(BINARY_NAME)
	test $(DEB_PACKAGE_NAME)
	test "$(DEB_PACKAGE_DESCRIPTION)"
	@ mkdir -p $(BUILD_ARTIFACTS_DIR)
	@ mkdir -p etc/$(BINARY_NAME)
	@ cp files/pkg/config/config.yaml etc/$(BINARY_NAME)/
	@ cp $(BINARY_NAME) $(BUILD_ARTIFACTS_DIR)
	fpm --output-type deb \
		--input-type dir \
		--chdir $(BUILD_ARTIFACTS_DIR) \
		--prefix /usr/bin \
		--name $(BINARY_NAME) \
		--version $(VERSION) \
		--description '$(DEB_PACKAGE_DESCRIPTION)' \
		--license 'Apache License 2.0' \
		--architecture x86_64 \
		--url https://github.com/Whyrl35/prometheus-saltstack-exporter \
		--maintainer "Ludovic Houdayer <ludovic@whyrl.fr>" \
		--config-files etc \
		--deb-upstream-changelog CHANGELOG.md \
		--deb-after-purge ./files/pkg/deb/deb-after-purge.sh \
		--deb-systemd files/pkg/systemd/$(BINARY_NAME).service \
		--deb-user prometheus \
		--deb-group prometheus \
		-p $(DEB_PACKAGE_NAME)-$(VERSION).deb \
		$(BINARY_NAME)
	@ mv *.deb $(BUILD_ARTIFACTS_DIR)
	@ rm -rf etc

## Need rpmbuild / fpm tools
build-rpm:			# Build RPM package (needs other tools) see: https://echorand.me/posts/building-golang-deb-packages/
	test $(BINARY_NAME)
	test $(RPM_PACKAGE_NAME)
	test "$(RPM_PACKAGE_DESCRIPTION)"
	fpm -s deb -t rpm $(BUILD_ARTIFACTS_DIR)/*.deb
	@ mv *.rpm $(BUILD_ARTIFACTS_DIR)

version:			# Display the current version
	@ echo $(VERSION)

tarball:			# Make a .tar.gz as a release
tarball: promu
	@ mkdir -p $(BUILD_ARTIFACTS_DIR)
	@ echo ">> building release tarball"
	@ $(PROMU) tarball
	@ mv *.tar.gz $(BUILD_ARTIFACTS_DIR)

promu:				# Dependencies on promu binary
	@GOOS=$(shell uname -s | tr A-Z a-z) \
		GOARCH=$(subst x86_64,amd64,$(patsubst i%86,386,$(shell uname -m))) \
		$(GO) get -u github.com/prometheus/promu

run:				# Run go run .
	@$(GO) run ./

clean:				# Clean unecessary files
	rm -f ./$(BINARY_NAME)
	rm -rf ./$(BUILD_ARTIFACTS_DIR)
	$(GO) mod tidy

help:				# Show this help
	@fgrep -h "#" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/#//'


.PHONY: all style format dependencies build test vet tarball run clean
