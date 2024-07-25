# If you update this file, please follow:
# https://suva.sh/posts/well-documented-makefiles/

.DEFAULT_GOAL:=help

# Go variables
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Output
TIME   = `date +%H:%M:%S`
GREEN  := $(shell printf "\033[32m")
RED    := $(shell printf "\033[31m")
CNone  := $(shell printf "\033[0m")
OK   = echo ${TIME} ${GREEN}[ OK ]${CNone}
ERR  = echo ${TIME} ${RED}[ ERR ]${CNone} "error:"

##@ Help Targets
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[0m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Static Analysis Targets

check-diff: reviewable ## Execute branch is clean
	git --no-pager diff
	git diff --quiet || ($(ERR) please run 'make reviewable' to include all changes && false)
	@$(OK) branch is clean

reviewable: fmt vet lint generate ## Ensure code is ready for review
	git submodule update --remote
	go mod tidy

fmt: ## Run go fmt against code
	go fmt ./...

vet: ## Run go vet against code
	go vet ./...

lint: golangci-lint ## Run golangci-lint against code
	$(GOLANGCI_LINT) run

generate:
	go generate ./...

##@ Test Targets
.PHONY: testacc
testacc: ## Run acceptance tests
	TF_ACC=1 go test -v $(TESTARGS) -covermode=atomic -coverpkg=./... -coverprofile=profile.cov ./... -timeout 120m

##@ Development Targets

DEV_PROVIDER_VERSION=100.100.100
dev-provider:  ## Generate dev provider
	bash generate_dev_provider.sh $(DEV_PROVIDER_VERSION)

# Tools Section

BIN_DIR ?= ./bin
bin-dir:
	test -d $(BIN_DIR) || mkdir $(BIN_DIR)

GOLANGCI_VERSION ?= 1.54.2
golangci-lint: bin-dir
	if ! test -f $(BIN_DIR)/golangci-lint-linux-amd64; then \
		curl -LOs https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_VERSION)/golangci-lint-$(GOLANGCI_VERSION)-linux-amd64.tar.gz; \
		tar -zxf golangci-lint-$(GOLANGCI_VERSION)-linux-amd64.tar.gz; \
		mv golangci-lint-$(GOLANGCI_VERSION)-*/golangci-lint $(BIN_DIR)/golangci-lint-linux-amd64; \
		chmod +x $(BIN_DIR)/golangci-lint-linux-amd64; \
		rm -rf ./golangci-lint-$(GOLANGCI_VERSION)-linux-amd64*; \
	fi
	if ! test -f $(BIN_DIR)/golangci-lint-$(GOOS)-$(GOARCH); then \
		curl -LOs https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_VERSION)/golangci-lint-$(GOLANGCI_VERSION)-$(GOOS)-$(GOARCH).tar.gz; \
		tar -zxf golangci-lint-$(GOLANGCI_VERSION)-$(GOOS)-$(GOARCH).tar.gz; \
		mv golangci-lint-$(GOLANGCI_VERSION)-*/golangci-lint $(BIN_DIR)/golangci-lint-$(GOOS)-$(GOARCH); \
		chmod +x $(BIN_DIR)/golangci-lint-$(GOOS)-$(GOARCH); \
		rm -rf ./golangci-lint-$(GOLANGCI_VERSION)-$(GOOS)-$(GOARCH)*; \
	fi
GOLANGCI_LINT=$(BIN_DIR)/golangci-lint-$(GOOS)-$(GOARCH)
