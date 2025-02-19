ARG builder_image
ARG builder_image_tag
FROM ${builder_image}:${builder_image_tag} AS builder

WORKDIR /jbpf-protobuf/build
RUN mkdir -p /root/go/src
RUN ln -s /jbpf-protobuf/pkg /root/go/src/jbpf_protobuf_cli
RUN cmake .. -DINITIALIZE_SUBMODULES=OFF && make jbpf_protobuf_cli
ENV PATH="$PATH:/jbpf-protobuf/out/bin"

ENTRYPOINT [ "/jbpf-protobuf/deploy/entrypoint.sh", "jbpf_protobuf_cli" ]
