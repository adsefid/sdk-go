.PHONY: deps fmt lint build
deps:
	go mod download
fmt:
	gofmt -l -w .
	goimports -l -w . 2>/dev/null || true
lint:
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run ./...; else echo "golangci-lint not installed — see README"; fi
build:
	go build ./...
