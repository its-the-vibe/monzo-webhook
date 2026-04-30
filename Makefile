.PHONY: test lint ci

test:
	go test ./...

lint:
	gofmt -l . | grep . && exit 1 || true
	go vet ./...

ci: lint test
