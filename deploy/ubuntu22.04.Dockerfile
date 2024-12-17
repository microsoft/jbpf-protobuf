FROM mcr.microsoft.com/mirror/docker/library/ubuntu:22.04

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

RUN apt -y install protobuf-compiler python3-pip curl
RUN wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz
ENV PATH="$PATH:/root/go/bin:/usr/local/go/bin"
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /root/go/bin v1.60.3

# Set the working directory and copy the project files
WORKDIR /jbpf-protobuf
COPY . /jbpf-protobuf

RUN pip3 install -r /jbpf-protobuf/3p/nanopb/requirements.txt
WORKDIR /jbpf-protobuf/build
RUN cmake .. && make -j

ENTRYPOINT [ "/jbpf-protobuf/out/bin/jbpf_protobuf_cli" ]