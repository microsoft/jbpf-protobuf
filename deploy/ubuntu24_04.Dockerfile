FROM mcr.microsoft.com/mirror/docker/library/ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive
SHELL ["/bin/bash", "-c"]
ENV CLANG_FORMAT_CHECK=1
ENV CPP_CHECK=1

RUN echo "*** Installing packages"
RUN apt update --fix-missing
RUN apt install -y cmake build-essential libboost-dev git libboost-program-options-dev \
    wget gcovr doxygen libboost-filesystem-dev libasan6 python3

RUN apt install -y clang-format cppcheck
RUN apt install -y clang gcc-multilib
RUN apt install -y libyaml-cpp-dev

RUN apt -y install protobuf-compiler python3-pip python3-protobuf python3-grpcio curl
RUN apt -y install golang-1.23 \
    golang-github-google-uuid-dev \
    golang-github-sirupsen-logrus-dev \
    golang-github-spf13-cobra-dev \
    golang-github-spf13-pflag-dev \
    golang-github-stretchr-testify-dev \
    golang-golang-x-sync-dev\
    golang-google-protobuf-dev \
    golang-gopkg-yaml.v3-dev

ENV PATH="$PATH:/root/go/bin:/usr/lib/go-1.23/bin:/usr/share/gocode/bin"
RUN go env -w GOFLAGS=-buildvcs=false
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.60.3
ENV GOPROXY=off
ENV GO111MODULE=off
ENV GOPATH="/root/go:/usr/lib/go-1.23:/usr/share/gocode"

# Set the working directory and copy the project files
WORKDIR /jbpf-protobuf
COPY . /jbpf-protobuf

ENTRYPOINT [ "/jbpf-protobuf/deploy/entrypoint.sh" ]
