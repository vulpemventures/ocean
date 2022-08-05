.PHONY: build build-cli clean cov docker-proto docker-proto-lint help proto proto-lint run test vet

## build: build for all platforms
build: 
	@echo "Building oceand binary..."
	@chmod u+x ./scripts/build
	@./scripts/build

## build-cli: build CLI for all platforms
build-cli: 
	@echo "Building ocean-cli binary..."
	@chmod u+x ./scripts/build-cli
	@./scripts/build-cli

clean:
	@echo "Cleaning..."
	@go clean

docker-proto-lint:
	@docker run --rm --volume "$(shell pwd):/workspace" --workdir /workspace bufbuild/buf lint

docker-proto: docker-proto-lint
	@docker run --rm --volume "$(shell pwd):/workspace" --workdir /workspace bufbuild/buf generate

proto: proto-lint
	@buf generate

proto-lint:
	@buf lint

fmt:
	@echo "Gofmt..."
	@if [ -n "$(gofmt -l .)" ]; then echo "Go code is not formatted"; exit 1; fi

run: clean
	@echo "Running oceand..."
	@export OCEAN_NETWORK=regtest; \
	export OCEAN_LOG_LEVEL=5; \
	export OCEAN_NO_TLS=true; \
	export OCEAN_STATS_INTERVAL=120; \
	export OCEAN_ESPLORA_URL=http://localhost:3001; \
	export OCEAN_NODE_PEERS="localhost:18886"; \
	export OCEAN_UTXO_EXPIRY_DURATION_IN_SECONDS=60; \
	go run ./cmd/oceand

test: fmt
	@echo "Testing..."
	go test --race --count=1 -v ./...

cov:
	@echo "Coverage..."
	go test -cover ./...

vet:
	@echo "Vet..."
	@go vet ./...

help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
