// Copyright (c) Microsoft Corporation. All rights reserved.
#define BOOST_BIND_GLOBAL_PLACEHOLDERS

#include <iostream>
#include <boost/property_tree/ptree.hpp>
#include <boost/property_tree/json_parser.hpp>
#include <string>
#include <cstring>
#include <csignal>

#include "schema.pb.h"

#include "jbpf.h"
#include "jbpf_hook.h"
#include "jbpf_defs.h"

using namespace std;

#define MAX_SERIALIZED_SIZE 1024

int sockfd;
struct sockaddr_in servaddr;

// Hook declaration and definition.
DECLARE_JBPF_HOOK(
    example,
    struct jbpf_generic_ctx ctx,
    ctx,
    HOOK_PROTO(packet *p, int ctx_id),
    HOOK_ASSIGN(ctx.ctx_id = ctx_id; ctx.data = (uint64_t)(void *)p; ctx.data_end = (uint64_t)(void *)(p + 1);))

DEFINE_JBPF_HOOK(example)

// Handler function that is invoked every time that jbpf receives one or more buffers of data from a codelet
static void
io_channel_forward_output(jbpf_io_stream_id_t *stream_id, void **bufs, int num_bufs, void *ctx)
{
    auto io_ctx = jbpf_get_io_ctx();
    if (io_ctx == NULL)
    {
        std::cerr << "Failed to get IO context. Got NULL" << std::endl;
        return;
    }

    char serialized[MAX_SERIALIZED_SIZE];
    int serialized_size;

    if (stream_id && num_bufs > 0)
    {
        // Fetch the data and print in JSON format
        for (auto i = 0; i < num_bufs; i++)
        {
            serialized_size = jbpf_io_channel_pack_msg(io_ctx, bufs[i], serialized, sizeof(serialized));
            if (serialized_size > 0)
            {
                sendto(sockfd, serialized, serialized_size,
                       MSG_CONFIRM, (const struct sockaddr *)&servaddr,
                       sizeof(servaddr));
                std::cout << "Message sent, size: " << serialized_size << std::endl;
            }
            else
            {
                std::cerr << "Failed to serialize message. Got return code: " << serialized_size << std::endl;
            }
        }
    }
}

bool done = false;

void sig_handler(int signo)
{
    done = true;
}

int handle_signal()
{
    if (signal(SIGINT, sig_handler) == SIG_ERR)
    {
        return 0;
    }
    if (signal(SIGTERM, sig_handler) == SIG_ERR)
    {
        return 0;
    }
    return -1;
}

void *fwd_socket_to_channel_in(void *arg)
{
    jbpf_register_thread();

    auto io_ctx = jbpf_get_io_ctx();

    if (io_ctx == NULL)
    {
        std::cerr << "Failed to get IO context. Got NULL" << std::endl;
        exit(0);
    }

    int sockfd, connfd;
    socklen_t len;
    struct sockaddr_in servaddr, cli;
    // socket create and verification
    sockfd = socket(AF_INET, SOCK_STREAM, 0);
    if (sockfd == -1)
    {
        printf("socket creation failed...\n");
        exit(0);
    }
    else
        printf("Socket successfully created..\n");
    bzero(&servaddr, sizeof(servaddr));
    servaddr.sin_family = AF_INET;
    servaddr.sin_addr.s_addr = htonl(INADDR_ANY);
    servaddr.sin_port = htons(20787);
    if ((bind(sockfd, (struct sockaddr *)&servaddr, sizeof(servaddr))) != 0)
    {
        printf("socket bind failed...\n");
        exit(0);
    }
    else
        printf("Socket successfully binded..\n");
    if ((listen(sockfd, 5)) != 0)
    {
        printf("Listen failed...\n");
        exit(0);
    }
    else
        printf("Server listening..\n");
    len = sizeof(cli);
    for (;;)
    {
        connfd = accept(sockfd, (struct sockaddr *)&cli, &len);
        if (connfd < 0)
        {
            printf("server accept failed...\n");
            exit(0);
        }
        else
            printf("server accept the client...\n");
        char buff[MAX_SERIALIZED_SIZE];
        int n;
        struct jbpf_io_stream_id stream_id = {0};
        for (;;)
        {
            auto n_diff = read(connfd, &buff[n], sizeof(buff) - n);
            n += n_diff;
            if (n_diff == 0)
            {
                printf("Client disconnected\n");
                break;
            }
            else if (n >= 18)
            {
                uint16_t payload_size = buff[1] * 256 + buff[0];
                if (n < payload_size + 2)
                {
                    continue;
                }
                else if (n > payload_size + 2)
                {
                    std::cerr << "Unexpected number of bytes in buffer, expected: " << payload_size << ", got: " << n - 2 << std::endl;
                    break;
                }

                jbpf_channel_buf_ptr deserialized = jbpf_io_channel_unpack_msg(io_ctx, &buff[2], payload_size, &stream_id);
                if (deserialized == NULL)
                {
                    std::cerr << "Failed to deserialize message. Got NULL" << std::endl;
                }
                else
                {
                    auto io_channel = jbpf_io_find_channel(io_ctx, stream_id, false);
                    if (io_channel)
                    {
                        auto ret = jbpf_io_channel_submit_buf(io_channel);
                        if (ret != 0)
                        {
                            std::cerr << "Failed to send message to channel. Got return code: " << ret << std::endl;
                        }
                        else
                        {
                            std::cout << "Dispatched msg of size: " << payload_size << std::endl;
                        }
                    }
                    else
                    {
                        std::cerr << "Failed to find io channel. Got NULL" << std::endl;
                    }
                }
                bzero(buff, MAX_SERIALIZED_SIZE);
                n = 0;
            }
        }
    }
    close(sockfd);
    // exit the current thread
    pthread_exit(NULL);
}

int main(int argc, char **argv)
{
    // Creating socket file descriptor
    if ((sockfd = socket(AF_INET, SOCK_DGRAM, 0)) < 0)
    {
        perror("socket creation failed");
        exit(EXIT_FAILURE);
    }

    memset(&servaddr, 0, sizeof(servaddr));

    // Filling server information
    servaddr.sin_family = AF_INET;
    servaddr.sin_port = htons(20788);
    servaddr.sin_addr.s_addr = INADDR_ANY;

    struct jbpf_config jbpf_config = {0};
    jbpf_set_default_config_options(&jbpf_config);

    // Enable LCM IPC interface using UNIX socket at the default socket path (the default is through C API)
    jbpf_config.lcm_ipc_config.has_lcm_ipc_thread = true;
    snprintf(
        jbpf_config.lcm_ipc_config.lcm_ipc_name,
        sizeof(jbpf_config.lcm_ipc_config.lcm_ipc_name) - 1,
        "%s",
        JBPF_DEFAULT_LCM_SOCKET);

    if (!handle_signal())
    {
        std::cout << "Could not register signal handler" << std::endl;
        return -1;
    }

    // Initialize jbpf
    if (jbpf_init(&jbpf_config) < 0)
    {
        return -1;
    }

    pthread_t ptid;
    pthread_create(&ptid, NULL, &fwd_socket_to_channel_in, NULL);

    // Any thread that calls a hook must be registered
    jbpf_register_thread();

    // Register the callback to handle output messages from codelets
    jbpf_register_io_output_cb(io_channel_forward_output);

    int i = 0;

    // Sample application code calling a hook every second
    while (!done)
    {
        packet p;
        p.seq_no = i;
        p.value = -i;

        std::stringstream ss;
        ss << "instance " << i;

        std::strcpy(p.name, ss.str().c_str());

        // Call hook and pass packet
        hook_example(&p, 1);
        sleep(1);
        i++;
    }

    jbpf_stop();
    pthread_cancel(ptid);
    exit(EXIT_SUCCESS);

    return 0;
}
