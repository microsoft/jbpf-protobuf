# Copyright (c) Microsoft Corporation. All rights reserved.
BINARY_NAME := jbpf_protobuf_cli
CURRENT_DIR = $(shell pwd)
NANO_PB ?= $(shell dirname $(shell pwd))/3p/nanopb
TEST_WORKDIR ?= $(shell dirname $(shell pwd))/testdata
REGENERATE_SNAPSHOT ?= false
OUT_DIR ?= .

.PHONY : format mod clean lint test testclean

${BINARY_NAME}: format clean mod
	CGO_ENABLED=0 go build --trimpath -o ${OUT_DIR}/${BINARY_NAME} main.go

mod:
	go mod tidy

clean:
	rm -f ${OUT_DIR}/${BINARY_NAME}

lint:
	golangci-lint run

format:
	@if command -v gofmt > /dev/null; then \
		echo "Running gofmt check..."; \
		gofmt -s -d .; \
	else \
		echo "gofmt not found, skipping..."; \
	fi

test:
	TEST_WORKDIR=${TEST_WORKDIR} \
		NANO_PB=${NANO_PB} \
		SNAPSHOT_DIR=${CURRENT_DIR}/__snapshots__ \
		REGENERATE_SNAPSHOT=${REGENERATE_SNAPSHOT} \
		go test -v ./...

testclean:
	rm -r ${CURRENT_DIR}/__snapshots__/*; \
	go clean -testcache
