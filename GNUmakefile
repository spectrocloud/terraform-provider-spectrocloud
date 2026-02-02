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

# Test variables
TEST ?= ./spectrocloud/...
TESTARGS ?=
TEST_COUNT ?= 1
PARALLEL ?= 4
ACCTEST_TIMEOUT ?= 120m
UNITTEST_TIMEOUT ?= 30m

.PHONY: test
test: ## Run unit tests only (excludes TestAcc* tests)
	@echo "Running unit tests..."
	go test $(TEST) -v -count=$(TEST_COUNT) -parallel=$(PARALLEL) \
		-run '^Test[^A]|^TestA[^c]|^TestAc[^c]' \
		-timeout $(UNITTEST_TIMEOUT) $(TESTARGS)

.PHONY: test-unit
test-unit: test ## Alias for 'test' target

.PHONY: testacc
testacc: ## Run acceptance tests (requires TF_ACC=1 and API credentials)
	@echo "Running acceptance tests..."
	TF_ACC=1 go test -v $(TESTARGS) -covermode=atomic -coverpkg=./... \
		-coverprofile=profile.cov $(TEST) -timeout $(ACCTEST_TIMEOUT)

.PHONY: testacc-vcr
testacc-vcr: ## Run acceptance tests with VCR replay mode
	@echo "Running acceptance tests with VCR replay..."
	TF_ACC=1 go test -v $(TEST) -run 'TestAcc' -timeout $(ACCTEST_TIMEOUT) $(TESTARGS)

.PHONY: testacc-vcr-record
testacc-vcr-record: ## Run acceptance tests in VCR record mode (requires API credentials)
	@echo "Running acceptance tests with VCR record mode..."
	@echo "WARNING: This will make real API calls and record them to cassette files"
	VCR_RECORD=true TF_ACC=1 go test -v $(TEST) -run 'TestAcc' -timeout $(ACCTEST_TIMEOUT) $(TESTARGS)

.PHONY: test-vcr
test-vcr: ## Run VCR-enabled unit tests
	@echo "Running VCR unit tests..."
	go test $(TEST) -v -run 'TestVCR' -timeout $(UNITTEST_TIMEOUT) $(TESTARGS)

.PHONY: test-vcr-record
test-vcr-record: ## Record VCR cassettes for unit tests (requires API credentials)
	@echo "Recording VCR cassettes..."
	VCR_RECORD=true go test $(TEST) -v -run 'TestVCR' -timeout $(UNITTEST_TIMEOUT) $(TESTARGS)

.PHONY: test-project
test-project: ## Run all project resource tests
	@echo "Running project resource tests..."
	go test ./spectrocloud/... -v -run 'Project' -timeout $(UNITTEST_TIMEOUT)

.PHONY: test-all
test-all: test testacc ## Run all tests (unit + acceptance)

##@ Development Targets
DEV_PROVIDER_VERSION=100.100.100
dev-provider:  ## Generate dev provider
	bash generate_dev_provider.sh $(DEV_PROVIDER_VERSION)

.PHONY: test-with-coverage
test-with-coverage: ## Show coverage from existing profile.cov
	TF_ACC=1 go test -v $(TESTARGS) -covermode=atomic -coverpkg=./... -coverprofile=profile.cov ./spectrocloud/... -timeout 120m
	@echo "Total coverage:"
	@go tool cover -func=profile.cov | grep total


.PHONY: coverage
coverage: ## Show coverage from existing profile.cov
	@go tool cover -func=profile.cov | grep total

# Tools Section

BIN_DIR ?= ./bin
bin-dir:
	test -d $(BIN_DIR) || mkdir $(BIN_DIR)

GOLANGCI_VERSION ?= 2.7.2
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
