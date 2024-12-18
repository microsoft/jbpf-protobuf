#!/bin/sh

$JBPFP_OUT_DIR/bin/jbpf_lcm_cli -u -c codeletset_unload_request.yaml

$JBPFP_OUT_DIR/bin/jbpf_protobuf_cli decoder unload -c codeletset_load_request.yaml
