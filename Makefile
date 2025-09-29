.PHONY: discord tfapi

generate:
	go generate ./...

check:
	golangci-lint run --timeout 3m ./...

fmt:
	golangci-lint fmt
