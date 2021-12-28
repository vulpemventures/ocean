.PHONY: proto doc

## proto: compile proto files
proto:
	chmod u+x ./scripts/compile_proto
	./scripts/compile_proto