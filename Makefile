GOLANGCI_LINT_VERSION ?= v2.13.2
# golangci-lint v2.13.x cannot read export data from Go 1.27.2 and newer.
LINT_TOOLCHAIN ?= go1.27.0

.PHONY: deps fmt lint build test
deps:
	go mod download
fmt:
	gofmt -l -w .
	goimports -l -w . 2>/dev/null || true
lint:
	go vet ./...
	GOTOOLCHAIN=$(LINT_TOOLCHAIN) go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...
build:
	go build ./...
test:
	go test ./...
