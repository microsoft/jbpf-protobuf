#!/bin/sh

set -e

$JBPFP_OUT_DIR/bin/jbpf_protobuf_cli decoder load -c codeletset_load_request.yaml

$JBPFP_OUT_DIR/bin/jbpf_lcm_cli -l -c codeletset_load_request.yaml
