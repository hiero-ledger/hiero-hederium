build: format lint
	go build ./...

test-unit: build
	go test ./test/unit/... -v

test-cover:
	go test -race -covermode=atomic -coverpkg=./internal/... -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

test-load:
	go test ./test/load/... -v || true

format:
	go fmt ./... || goimports -w .

lint:
	golangci-lint run

set-previewnet:
	echo export CONFIG_FILE=""

set-testnet:
	echo export CONFIG_FILE=""

