ARG builder_image
ARG builder_image_tag
FROM ${builder_image}:${builder_image_tag} AS builder

WORKDIR /jbpf-protobuf/build
RUN cmake .. && make jbpf_protobuf_cli

FROM mcr.microsoft.com/azurelinux/distroless/base:3.0
COPY --from=builder /jbpf-protobuf/out/bin/jbpf_protobuf_cli /usr/local/bin/jbpf_protobuf_cli

ENTRYPOINT [ "jbpf_protobuf_cli" ]
