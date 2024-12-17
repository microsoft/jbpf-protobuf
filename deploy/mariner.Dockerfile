FROM mcr.microsoft.com/azurelinux/base/core:3.0

RUN echo "*** Installing packages"
RUN tdnf upgrade tdnf --refresh -y
RUN tdnf -y update
RUN tdnf -y install build-essential cmake git
RUN tdnf -y install yaml-cpp-devel yaml-cpp-static boost-devel gcovr clang python3
RUN tdnf -y install doxygen
## clang-format
RUN tdnf -y install clang-tools-extra

RUN tdnf -y install golang ca-certificates jq protobuf python3-pip
ENV PATH="$PATH:/root/go/bin"
RUN go env -w GOFLAGS=-buildvcs=false
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /root/go/bin v1.60.3

WORKDIR /jbpf-protobuf
COPY . /jbpf-protobuf

RUN pip3 install -r /jbpf-protobuf/3p/nanopb/requirements.txt
