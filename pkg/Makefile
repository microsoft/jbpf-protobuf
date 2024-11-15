# Copyright (c) Microsoft Corporation. All rights reserved.
BINARY_NAME := jbpf_protobuf_cli
CURRENT_DIR = $(shell pwd)
NANO_PB ?= $(shell dirname $(shell pwd))/3p/nanopb
TEST_WORKDIR ?= $(shell dirname $(shell pwd))/testdata
REGENERATE_SNAPSHOT ?= false
OUT_DIR ?= .

.PHONY : mod clean lint test testclean

${BINARY_NAME}: clean mod
	CGO_ENABLED=0 go build --trimpath -o ${OUT_DIR}/${BINARY_NAME} main.go

mod:
	go mod tidy

clean:
	rm -f ${OUT_DIR}/${BINARY_NAME}

lint:
	golangci-lint run

test:
	TEST_WORKDIR=${TEST_WORKDIR} \
		NANO_PB=${NANO_PB} \
		SNAPSHOT_DIR=${CURRENT_DIR}/__snapshots__ \
		REGENERATE_SNAPSHOT=${REGENERATE_SNAPSHOT} \
		go test -v ./...

testclean:
	rm -r ${CURRENT_DIR}/__snapshots__/*; \
	go clean -testcache
