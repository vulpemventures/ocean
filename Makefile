.PHONY: docker-lint docker-proto lint proto

docker-lint:
	@docker run --rm --volume "$(shell pwd):/workspace" --workdir /workspace bufbuild/buf lint

docker-proto: docker-lint
	@docker run --rm --volume "$(shell pwd):/workspace" --workdir /workspace bufbuild/buf generate

lint:
	@buf lint

proto: lint
	@buf generate