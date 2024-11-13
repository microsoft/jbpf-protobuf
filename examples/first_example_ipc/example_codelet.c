// Copyright (c) Microsoft Corporation. All rights reserved.

#include <string.h>

#include "jbpf_defs.h"
#include "jbpf_helper.h"
#include "schema.pb.h"

// Output map of type JBPF_MAP_TYPE_RINGBUF.
// The map is used to send out data of type packet.
// It holds a ringbuffer with a total of 3 elements.
jbpf_ringbuf_map(outmap, packet, 3);

// Input map of type JBPF_MAP_TYPE_CONTROL_INPUT.
// The map is used to receive data of type manual_ctrl_event.
// It uses a ringbuffer, that can store a total of 3 elements.
jbpf_control_input_map(inmap, manual_ctrl_event, 3);

// A map of type JBPF_MAP_TYPE_ARRAY, which is used
// to store internal codelet state.
struct jbpf_load_map_def SEC("maps") counter = {
    .type = JBPF_MAP_TYPE_ARRAY,
    .key_size = sizeof(int),
    .value_size = sizeof(int),
    .max_entries = 1,
};

SEC("jbpf_generic")
uint64_t
jbpf_main(void *state)
{

    void *c;
    int cnt;
    struct jbpf_generic_ctx *ctx;
    packet *p, *p_end;
    packet echo;
    manual_ctrl_event resp = {0};
    uint64_t index = 0;

    ctx = state;

    c = jbpf_map_lookup_elem(&counter, &index);
    if (!c)
        return 1;

    cnt = *(int *)c;
    cnt++;
    *(uint32_t *)c = cnt;

    p = (packet *)ctx->data;
    p_end = (packet *)ctx->data_end;

    if (p + 1 > p_end)
        return 1;

    echo = *p;

    // Copy the data that was passed to the codelet to the outmap ringbuffer
    // and send them out.
    if (jbpf_ringbuf_output(&outmap, &echo, sizeof(echo)) < 0)
    {
        return 1;
    }

    if (jbpf_control_input_receive(&inmap, &resp, sizeof(resp)) == 1)
    {
        // Print a debug message. This helper function should NOT be used in production environments, due to
        // its performance overhead. The helper function will be ignored, if *jbpf* has been built with the
        // USE_JBPF_PRINTF_HELPER option set to OFF.
        jbpf_printf_debug(" Called %d times so far and received manual_ctrl_event with value %d\n\n", cnt, resp.value);
    }

    return 0;
}
