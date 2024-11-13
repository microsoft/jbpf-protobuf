# jbpf-protobuf

This repository is a extension for [jbpf](https://github.com/microsoft/jbpf/) demonstrating how to utilize protobuf serialization as part of jbpf.

Prerequisites:
* C compiler
* Go v1.23.2+
* Make
* Pip
* Python

The project utilizes [Nanopb](https://github.com/nanopb/nanopb) to generate C structures for given protobuf specs that use contiguous memory. It also generates serializer libraries that can be provided to jbpf, to encode output and decode input data to seamlessly integrate external data processing systems.

# Getting started

```sh
# init submodules:
./init_submodules.sh

# source environment variables
source ./setup_jbpfp_env.sh

# build jbpf_protobuf_cli
make -C pkg
```

Alternatively, build using a container:
```sh
# init submodules:
./init_submodules.sh

docker build -t jbpf_protobuf_builder:latest -f deploy/Dockerfile .
```

## Running the examples

In order to run any of the samples, you'll need to build Janus.

```sh
mkdir -p jbpf/build
cd jbpf/build
cmake .. -DJBPF_EXPERIMENTAL_FEATURES=on
make -j
cd ../..
```

Then follow [these](./examples/first_example_standalone/README.md) steps to run a simple example.

# License

The jbpf framework is licensed under the [MIT license](LICENSE.md).
