#!/bin/bash

rm -rf 3p/nanopb jbpf
git submodule update --init --recursive
cd jbpf
./init_and_patch_submodules.sh
cd ..