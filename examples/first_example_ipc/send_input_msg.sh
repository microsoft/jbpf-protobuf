#!/bin/sh

$JBPFP_PATH/pkg/jbpf_protobuf_cli input forward \
  -c codeletset_load_request.yaml \
  --stream-id 11111111-1111-1111-1111-111111111111 \
  --inline-json "{\"value\": $1}"
