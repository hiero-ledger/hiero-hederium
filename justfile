build: format lint
	go build ./...

test-unit: build
	go test ./test/unit/... -v

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

