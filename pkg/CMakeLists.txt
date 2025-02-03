cmake_minimum_required(VERSION 3.16)
project(jbpf_protobuf_cli)

# Set the Go executable
find_program(GO_EXECUTABLE go)

if(NOT GO_EXECUTABLE)
    message(FATAL_ERROR "Go compiler not found. Please install Go.")
endif()

# Set variables
set(BINARY_NAME jbpf_protobuf_cli)
set(CURRENT_DIR ${CMAKE_CURRENT_SOURCE_DIR})
set(NANO_PB ${CMAKE_CURRENT_SOURCE_DIR}/../3p/nanopb)
set(TEST_WORKDIR ${CMAKE_CURRENT_SOURCE_DIR}/../testdata)
option(JBPFPCLI_REGENERATE_SNAPSHOT false)

# Add a custom target to build the Go application
add_custom_target(${BINARY_NAME} ALL
    COMMAND ${GO_EXECUTABLE} mod tidy
    COMMAND ${GO_EXECUTABLE} build --trimpath -o ${OUTPUT_DIR}/bin/${BINARY_NAME} main.go
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    COMMENT "Building jbpf_protobuf_cli application"
)

# Add a custom target to run linting
find_program(GOLANGCI_LINT_EXECUTABLE golangci-lint)

if(NOT GOLANGCI_LINT_EXECUTABLE)
    message(FATAL_ERROR "golangci-lint not found. Please install golangci-lint.")
endif()

add_custom_target(${BINARY_NAME}-lint ALL
    COMMAND ${GOLANGCI_LINT_EXECUTABLE} run
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    COMMENT "Running linting for jbpf_protobuf_cli"
)

# Add a custom target to run tests
add_custom_target(${BINARY_NAME}-test ALL
    COMMAND ${CMAKE_COMMAND} -E env
        TEST_WORKDIR=${TEST_WORKDIR}
        NANO_PB=${NANO_PB}
        SNAPSHOT_DIR=${CURRENT_DIR}/__snapshots__
        REGENERATE_SNAPSHOT=${JBPFPCLI_REGENERATE_SNAPSHOT}
        ${GO_EXECUTABLE} test -v ./...
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    COMMENT "Running tests for jbpf_protobuf_cli"
)

# Add a custom target to run tests
find_program(GOFMT_EXECUTABLE gofmt)

if(NOT GOFMT_EXECUTABLE)
    message(FATAL_ERROR "gofmt not found. Please install gofmt.")
endif()

add_custom_target(${BINARY_NAME}-format
    COMMAND ${GOFMT_EXECUTABLE} -s -d .
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    COMMENT "Running formatting for jbpf_protobuf_cli"
)
