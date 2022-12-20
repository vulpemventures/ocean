# Ocean

This repo contains the protobuf files of the Ocean wallet interface.  
Any Ocean wallet implementation must stick with the services and RPCs defined by the protos.  
This also includes a single-key Ocean wallet that can be served by running the binary or as a dockerized solution.

We use Buf as package manager for the protos, you should import them from the [registry](https://buf.build/equitas-foundation/ocean) and compile the stubs for your preferred prorgramming language with buf CLI.

## Build

Build ocean binaries:

```bash
# build oceand
$ make build

# build ocean CLI
$ make build-cli
```

Build docker image:

```bash
$ docker build -t ghcr.io/equitas-foundation/oceand:latest .
```

## Local run

```bash
# run oceand with regtest configuration
$ make run

# in another tab, check the status of the daemon with the CLI
$ alias ocean=$(pwd)/build/ocean-cli-<os>-<arch>
$ ocean config init --no-tls
$ ocean wallet status
# check all available commands with help message
$ ocean --help
```

## Test

```bash
# run unit and compose tests:
$ make test
```

## Release

Precompiled binaries are published with each [release](https://github.com/equitas-foundation/bamp-ocean/releases).

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the
[tags on this repository](https://github.com/equitas-foundation/bamp-ocean/tags). 

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.