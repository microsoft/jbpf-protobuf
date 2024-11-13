# Copyright (c) Microsoft Corporation. All rights reserved.
ifeq ($(BUILD_TYPE),Debug)
    DEBUG_CFLAGS = -g
	DEBUG_LDFLAGS = -lgcov
else ifeq ($(BUILD_TYPE),AddressSanitizer)
	DEBUG_CFLAGS = -fsanitize=address
endif

AGENT_NAME := example_app
CODELET_NAME := example_codelet.o
INCLUDES := -I${JBPF_OUT_DIR}/inc -I${NANO_PB} -DJBPF_EXPERIMENTAL_FEATURES=on
LDFLAGS := -L${JBPF_OUT_DIR}/lib -ljbpf -lck -lubpf -lmimalloc -lpthread -ldl -lrt ${DEBUG_LDFLAGS}
AGENT_FILE := example_app.cpp
CODELET_FILE := example_codelet.c
CODELET_CC := clang
JBPF_PROTOBUF_CLI := ${JBPFP_PATH}/pkg/jbpf_protobuf_cli

CODELET_CFLAGS := -O2 -target bpf -Wall -DJBPF_DEBUG_ENABLED -D__x86_64__

.PHONY: all clean

all: clean schema codelet agent

codelet: ${CODELET_FILE}
	${CODELET_CC} ${CODELET_CFLAGS} ${INCLUDES} -c ${CODELET_FILE} -o ${CODELET_NAME}

schema:
	${JBPF_PROTOBUF_CLI} serde -s schema:packet,manual_ctrl_event; \
	rm -f *_serializer.c

agent:
	g++ -std=c++17 $(INCLUDES) -o ${AGENT_NAME} $(AGENT_FILE) ${DEBUG_CFLAGS} ${LDFLAGS}

clean:
	rm -f ${AGENT_NAME} ${CODELET_NAME} *.pb.h *.pb.c *.pb *.so

