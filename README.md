# jbpf-protobuf
[![jbpf-protobuf build](https://github.com/microsoft/jbpf-protobuf/actions/workflows/workflow.yaml/badge.svg?branch=main)](https://github.com/microsoft/jbpf-protobuf/actions/workflows/workflow.yaml)

**NOTE: This project uses an experimental feature from jbpf. It is not meant to be used in production environments.**

This repository is a extension for [jbpf](https://github.com/microsoft/jbpf/) demonstrating how to utilize protobuf serialization as part of jbpf.

Prerequisites:
* C compiler
* Go v1.23.2+
* Make
* Pip
* Python3
* Protocol Buffer Compiler (protoc)

The project utilizes [Nanopb](https://github.com/nanopb/nanopb) to generate C structures for given protobuf specs that use contiguous memory. It also generates serializer libraries that can be provided to jbpf, to encode output and decode input data to seamlessly integrate external data processing systems.

# Getting started

```sh
# Install nanopb pip packages:
python3 -m pip install -r 3p/nanopb/requirements.txt

# source environment variables
source ./setup_jbpfp_env.sh

# build cli and dependencies
mkdir build
cd build
cmake .. && make -j
```

Alternatively, build using a container:
```sh
# init submodules:
./init_submodules.sh

# Create builder image with all dependencies loaded
OS=azurelinux # see ./deploy directory for currently supported OS versions
docker build -t jbpfp-$OS:latest -f deploy/$OS.Dockerfile .

# Build the cli and dependencies
docker run --rm -it \
  -v $(pwd):/jbpf-protobuf \
  -w /jbpf-protobuf/build \
  jbpfp-$OS:latest \
  cmake -DINITIALIZE_SUBMODULES=off ..

docker run --rm -it \
  -v $(pwd):/jbpf-protobuf \
  -w /jbpf-protobuf/build \
  jbpfp-$OS:latest make -j
```

## Running the examples

Once the project is built you can run the sample apps. Follow [these](./examples/first_example_standalone/README.md) steps to run a simple example.

# License

The jbpf framework is licensed under the [MIT license](LICENSE.md).
