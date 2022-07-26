DEV_PROVIDER_VERSION=100.100.100

default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

dev-provider:
	bash generate_dev_provider.sh $(DEV_PROVIDER_VERSION)