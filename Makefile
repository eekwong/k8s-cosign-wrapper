UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	LDFLAGS = -extldflags "-static"
endif

GO_BUILDER_VERSION = 1.19.3
COSIGN_VERSION = 1.13.1

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: clean
clean: ## Clean all artifacts
	@echo "+ $@"
	rm -fr dist

.PHONY: run
run: ## Run k8s-cosign-wrapper
	go run github.com/eekwong/k8s-cosign-wrapper/cmd/k8s-cosign-wrapper

.PHONY: binary
binary: ## Build k8s-cosign-wrapper go binary
	@echo "+ $@"
	go build -a \
		-ldflags "$(LDFLAGS)" \
		-o dist/k8s-cosign-wrapper github.com/eekwong/k8s-cosign-wrapper/cmd/k8s-cosign-wrapper

.PHONY: image
image: ## Build k8s-cosign-wrapper docker image
	@echo "+ $@"
	docker build \
		--build-arg COSIGN_VERSION=$(COSIGN_VERSION) \
		--build-arg GO_BUILDER_VERSION=$(GO_BUILDER_VERSION) \
		-f build/Dockerfile \
		-t matthewkwong/k8s-cosign-wrapper:$(COSIGN_VERSION) ./
