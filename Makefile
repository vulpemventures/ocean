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
	export OCEAN_NODE_RPC_ADDR=http://admin1:123@localhost:18884; \
	export OCEAN_UTXO_EXPIRY_DURATION_IN_SECONDS=60; \
	go run ./cmd/oceand

test: fmt pgcreatetestdb
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

#### Postgres database ####
## pg: starts postgres db inside docker container
pg:
	docker run --name oceand-pg -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres

## droppg: stop and remove postgres container
droppg:
	docker stop oceand-pg
	docker rm oceand-pg

## createdb: create db inside docker container
createdb:
	docker exec oceand-pg createdb --username=root --owner=root oceand-db

## dropdb: drops db inside docker container
dropdb:
	docker exec oceand-pg dropdb oceand-db

## createtestdb: create test db inside docker container
createtestdb:
	docker exec oceand-pg createdb --username=root --owner=root oceand-db-test

## droptestdb: drops test db inside docker container
droptestdb:
	docker exec oceand-pg dropdb oceand-db-test

## recreatedb: drop and create main and test db
recreatedb: dropdb createdb droptestdb createtestdb

## recreatetestdb: drop and create test db
recreatetestdb: droptestdb createtestdb

## pgcreatetestdb: starts docker container and creates test db, used in CI
pgcreatetestdb:
	chmod u+x ./scripts/create_testdb
	./scripts/create_testdb

## psql: connects to postgres terminal running inside docker container
psql:
	docker exec -it oceand-pg psql -U root -d oceand-db

## migrate: creates pg migration file(eg. make FILE=init migrate)
migrate:
	migrate create -ext sql -dir ./internal/infrastructure/storage/db/postgres/migration/ $(FILE)

## sqlc: gen sql
sqlc:
	@echo "gen sql..."
	cd internal/infrastructure/storage/db/postgres; sqlc generate