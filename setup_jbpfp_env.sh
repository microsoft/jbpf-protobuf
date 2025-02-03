#!/bin/bash

export JBPFP_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd)"
export NANO_PB=$JBPFP_PATH/3p/nanopb
export JBPFP_OUT_DIR=$JBPFP_PATH/out
source $JBPFP_PATH/jbpf/setup_jbpf_env.sh
