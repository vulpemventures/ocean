.PHONY: proto

proto:
	docker run --volume "$(shell pwd)/api-spec/protobuf:/specs" --workdir /specs bufbuild/buf generate