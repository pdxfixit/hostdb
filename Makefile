default: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s %s\n\033[0m", $$1, $$2}'

.PHONY: get
get: ## go get will ensure dependencies are present
	go get -d

.PHONY: fmt
fmt: ## go fmt
	go fmt ./...

.PHONY: vet
vet: ## go vet
	go vet -v ./...

.PHONY: lint
lint: ## golint
	go get golang.org/x/lint/golint
	go list ./... | xargs -L1 golint

.PHONY: errcheck
errcheck: ## errcheck
	go get github.com/kisielk/errcheck
	errcheck ./...

.PHONY: test
test: fmt vet lint errcheck ## format, vet, lint, errorcheck, then run tests and cleanup
	go test -v --failfast
	rm -rf ./sample-data
