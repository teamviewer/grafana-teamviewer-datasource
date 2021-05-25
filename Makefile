GO                      ?= GO111MODULE=on go
GOPATH                  := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
GOLINTER                ?= $(GOPATH)/bin/golangci-lint
pkgs                    = $(shell $(GO) list ./... | grep -v /vendor/)
PREFIX                  ?= $(shell pwd)
PACKAGE_NAME            := "teamviewer-datasource"

.PHONY: all
all: clean build-plugin-backend

.PHONY: build-plugin-frontend
build-plugin-frontend:
	@echo ">> Build plugin frontend"
	@yarn install
	@yarn build

.PHONY: update-grafana-sdk
update-grafana-sdk:
	@echo ">> Update Grafana plugin SDK"
	@$(GO) install github.com/grafana/grafana-plugin-sdk-go

.PHONY: build-pluging-backend
build-plugin-backend: build-plugin-frontend update-grafana-sdk
	@echo ">> Build plugin backend"
	@mage -v

.PHONY: sign
sign:
	@echo ">> Sign built Grafana Plugin"
ifeq ($(strip $(GRAFANA_API_KEY)),)
	@echo "Environment variable GRAFANA_API_KEY missing!!";exit 1;
endif
	@npx @grafana/toolkit plugin:sign

.PHONY: linting
linting: $(GOLINTER)
	@echo ">> linting code"
	@$(GOLINTER) run --config $(PREFIX)/.golangci.yml

.PHONY: lint
$(GOPATH)/bin/golangci-lint lint:
	@GOOS=$(shell uname -s | tr A-Z a-z) \
		GOARCH=$(subst x86_64,amd64,$(patsubst i%86,386,$(shell uname -m))) \
		$(GO) get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.40.1

.PHONY: clean
clean:
	@echo ">> clean node_modules"
	@rm -rf node_modules
	@echo ">> clean binaries"
	@rm -rf "$(PACKAGE_NAME)" dist vendor "$(PACKAGE_NAME).zip"
