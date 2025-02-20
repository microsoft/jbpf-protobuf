FROM mcr.microsoft.com/azurelinux/base/core:3.0

RUN echo "*** Installing packages"
RUN tdnf upgrade tdnf --refresh -y
RUN tdnf -y update
RUN tdnf -y install build-essential cmake git
RUN tdnf -y install yaml-cpp-devel yaml-cpp-static boost-devel gcovr clang python3
RUN tdnf -y install doxygen
## clang-format
RUN tdnf -y install clang-tools-extra

RUN tdnf -y install golang ca-certificates jq protobuf python3-grpcio
RUN go env -w GOFLAGS=-buildvcs=false
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.60.3
ENV PATH="/root/go/bin:${PATH}"

WORKDIR /jbpf-protobuf
COPY . /jbpf-protobuf

RUN pip3 install -r /jbpf-protobuf/3p/nanopb/requirements.txt
## TODO: We prefer the following rather than above
## But the protobuf 3.20 is a hard requirement which is not available on ubuntu 22.04 through apt.
## RUN tdnf install -y python3-protobuf python3-grpcio

ENTRYPOINT [ "/jbpf-protobuf/deploy/entrypoint.sh" ]
