#!/bin/sh

set -e

$JBPFP_PATH/pkg/jbpf_protobuf_cli decoder load -c codeletset_load_request.yaml

$JBPF_PATH/out/bin/jbpf_lcm_cli -l -c codeletset_load_request.yaml
