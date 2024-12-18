#!/bin/sh

$JBPFP_OUT_DIR/bin/jbpf_protobuf_cli input forward \
  -c codeletset_load_request.yaml \
  --stream-id 11111111-1111-1111-1111-111111111111 \
  --inline-json "{\"value\": $1}"
