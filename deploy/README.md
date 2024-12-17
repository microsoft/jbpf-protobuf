# Images

The Dockerfiles in this directory can be split into two categories:
1. builder images for different operating systems:
  * [mariner](./mariner.Dockerfile)
  * [Ubuntu20.04](./ubuntu20_04.Dockerfile)
  * [Ubuntu22.04](./ubuntu22_04.Dockerfile)
2. Containerized application
  * [jbpf_protobuf_cli](./jbpf_protobuf_cli.Dockerfile) is built from one of the OS images in step 1.

## Usage

```sh
# To build for a particular OS, run:
OS=mariner
docker build -t jbpfp-$OS:latest -f deploy/$OS.Dockerfile .

# And to create a jbpf_protobuf_cli image from that container, run:
docker build --build-arg builder_image=jbpfp-$OS --build-arg builder_image_tag=latest -t jbpf_protobuf_cli:latest - < deploy/jbpf_protobuf_cli.Dockerfile
```