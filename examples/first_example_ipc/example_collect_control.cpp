#include <iostream>
#include <unistd.h>

#include "jbpf.h"
#include "jbpf_defs.h"
#include "jbpf_io.h"
#include "jbpf_io_channel.h"
#include <signal.h>

#include "schema.pb.h"

#define SHM_NAME "example_ipc_app"

#define MAX_SERIALIZED_SIZE 1024

int sockfd;
struct sockaddr_in servaddr;
static volatile int done = 0;

static void
handle_channel_bufs(
    struct jbpf_io_channel* io_channel, struct jbpf_io_stream_id* stream_id, void** bufs, int num_bufs, void* ctx)
{
    struct jbpf_io_ctx* io_ctx = static_cast<struct jbpf_io_ctx*>(ctx);
    char serialized[MAX_SERIALIZED_SIZE];
    int serialized_size;

    if (stream_id && num_bufs > 0) {
        // Fetch the data and send to local decoder
        for (auto i = 0; i < num_bufs; i++) {
            serialized_size = jbpf_io_channel_pack_msg(io_ctx, bufs[i], serialized, sizeof(serialized));
            if (serialized_size > 0) {
                sendto(
                    sockfd,
                    serialized,
                    serialized_size,
                    MSG_CONFIRM,
                    (const struct sockaddr*)&servaddr,
                    sizeof(servaddr));
                std::cout << "Message sent, size: " << serialized_size << std::endl;
            } else {
                std::cerr << "Failed to serialize message. Got return code: " << serialized_size << std::endl;
            }

	    jbpf_io_channel_release_buf(bufs[i]);
        }
    }
}

void*
fwd_socket_to_channel_in(void* arg)
{
    struct jbpf_io_ctx* io_ctx = static_cast<struct jbpf_io_ctx*>(arg);

    jbpf_io_register_thread();

    int sockfd, connfd;
    socklen_t len;
    struct sockaddr_in servaddr, cli;

    // socket create and verification
    sockfd = socket(AF_INET, SOCK_STREAM, 0);
    if (sockfd == -1) {
        printf("socket creation failed...\n");
        exit(0);
    } else
        printf("Socket successfully created..\n");
    bzero(&servaddr, sizeof(servaddr));

    servaddr.sin_family = AF_INET;
    servaddr.sin_addr.s_addr = htonl(INADDR_ANY);
    servaddr.sin_port = htons(20787);

    if ((bind(sockfd, (struct sockaddr*)&servaddr, sizeof(servaddr))) != 0) {
        printf("socket bind failed...\n");
        exit(0);
    } else
        printf("Socket successfully binded..\n");

    if ((listen(sockfd, 5)) != 0) {
        printf("Listen failed...\n");
        exit(0);
    } else
        printf("Server listening..\n");
    len = sizeof(cli);

    for (;;) {
        connfd = accept(sockfd, (struct sockaddr*)&cli, &len);
        if (connfd < 0) {
            printf("server accept failed...\n");
            exit(0);
        } else
            printf("server accept the client...\n");

        char buff[MAX_SERIALIZED_SIZE];
        int n;
        struct jbpf_io_stream_id stream_id = {0};

        for (;;) {
            auto n_diff = read(connfd, &buff[n], sizeof(buff) - n);
            n += n_diff;
            if (n_diff == 0) {
                printf("Client disconnected\n");
                break;
            } else if (n >= 18) {
                uint16_t payload_size = buff[1] * 256 + buff[0];
                if (n < payload_size + 2) {
                    continue;
                } else if (n > payload_size + 2) {
                    std::cerr << "Unexpected number of bytes in buffer, expected: " << payload_size
                              << ", got: " << n - 2 << std::endl;
                    break;
                }

                jbpf_channel_buf_ptr deserialized =
                    jbpf_io_channel_unpack_msg(io_ctx, &buff[2], payload_size, &stream_id);
                if (deserialized == NULL) {
                    std::cerr << "Failed to deserialize message. Got NULL" << std::endl;
                } else {
                    auto io_channel = jbpf_io_find_channel(io_ctx, stream_id, false);
                    if (io_channel) {
                        auto ret = jbpf_io_channel_submit_buf(io_channel);
                        if (ret != 0) {
                            std::cerr << "Failed to send message to channel. Got return code: " << ret << std::endl;
                        } else {
                            std::cout << "Dispatched msg of size: " << payload_size << std::endl;
                        }
                    } else {
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

void
handle_ctrl_c(int signum)
{
    printf("\nCaught Ctrl+C! Exiting gracefully...\n");
    done = 1;
}

int
main(int argc, char** argv)
{
    signal(SIGINT, handle_ctrl_c);

    // Creating socket file descriptor
    if ((sockfd = socket(AF_INET, SOCK_DGRAM, 0)) < 0) {
        perror("socket creation failed");
        exit(EXIT_FAILURE);
    }

    memset(&servaddr, 0, sizeof(servaddr));

    // Filling server information
    servaddr.sin_family = AF_INET;
    servaddr.sin_port = htons(20788);
    servaddr.sin_addr.s_addr = INADDR_ANY;

    struct jbpf_io_config io_config = {0};
    struct jbpf_io_ctx* io_ctx;

    // Designate the data collection framework as a primary for the IPC
    io_config.type = JBPF_IO_IPC_PRIMARY;

    strncpy(io_config.ipc_config.addr.jbpf_io_ipc_name, SHM_NAME, JBPF_IO_IPC_MAX_NAMELEN);

    // Configure memory size for the IO buffer
    io_config.ipc_config.mem_cfg.memory_size = JBPF_HUGEPAGE_SIZE_1GB;

    // Configure the jbpf agent to operate in shared memory mode
    io_ctx = jbpf_io_init(&io_config);

    if (!io_ctx) {
        return -1;
    }

    pthread_t ptid;
    pthread_create(&ptid, NULL, &fwd_socket_to_channel_in, io_ctx);

    // Every thread that sends or receives jbpf data needs to be registered using this call
    jbpf_io_register_thread();

    while (!done) {
        // Continuously poll IPC output buffers
        jbpf_io_channel_handle_out_bufs(io_ctx, handle_channel_bufs, io_ctx);
        sleep(1);
    }

    pthread_cancel(ptid);
    return 0;
}
