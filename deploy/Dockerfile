FROM mcr.microsoft.com/oss/go/microsoft/golang:1.23.2-1-azurelinux3.0 AS builder

RUN tdnf upgrade tdnf --refresh -y && tdnf -y update
RUN tdnf install -y make python3-pip awk jq
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /root/go/bin v1.60.3
ENV PATH="$PATH:/root/go/bin"

COPY pkg /workspace/pkg
COPY 3p /workspace/3p
RUN python3 -m pip install -r /workspace/3p/nanopb/requirements.txt
COPY testdata /workspace/testdata
ENV NANO_PB=/workspace/3p/nanopb

RUN make -C /workspace/pkg
RUN make -C /workspace/pkg test lint -j

FROM mcr.microsoft.com/azurelinux/base/core:3.0
RUN tdnf upgrade tdnf --refresh -y && tdnf -y update
RUN tdnf install -y build-essential make python3-pip

COPY --from=builder /workspace/3p/nanopb /nanopb
RUN python3 -m pip install -r /nanopb/requirements.txt
COPY --from=builder /workspace/pkg/jbpf_protobuf_cli /usr/local/bin/jbpf_protobuf_cli
ENV NANO_PB=/nanopb

ENTRYPOINT [ "jbpf_protobuf_cli" ]
