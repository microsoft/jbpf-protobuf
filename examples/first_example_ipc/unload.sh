#!/bin/sh

$JBPF_PATH/out/bin/jbpf_lcm_cli -u -c codeletset_unload_request.yaml

$JBPFP_PATH/pkg/jbpf_protobuf_cli decoder unload -c codeletset_load_request.yaml
