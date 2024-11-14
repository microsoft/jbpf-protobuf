# Basic example of standalone *jbpf* operation

This example showcases a basic *jbpf-protobuf* usage scenario, when using in IPC mode. It provides a C++ application (`example_collect_control`)
 that initializes *jbpf* in IPC primary mode, a dummy C++ application (`example_app`), that initializes
*jbpf* in IPC secondary mode, and an example codelet (`example_codelet.o`).
The example demonstrates the following:
1. How to declare and call hooks in the *jbpf* secondary process.
2. How to collect data sent by the codelet from the *jbpf* primary process.
3. How to forward data sent by the codelet onwards to a local decoder using a UDP socket.
4. How to receive data sent by the decoder using a TCP socket onwards to the primary process.
5. How to load and unload codeletsets using the LCM CLI tool (via a Unix socket API).

For more details of the exact behavior of the application and the codelet, you can check the inline comments in [example_collect_control.cpp](./example_collect_control.cpp),
[example_app.cpp](./example_app.cpp) and [example_codelet.c](./example_codelet.c)

## Usage

This example expects *jbpf* to be built (see [README.md](../../README.md)).

To build the example from scratch, we run the following commands:
```sh
$ source ../../setup_jbpfp_env.sh
$ make
```

This should produce these artifacts:
* `example_collect_control`
* `example_app`
* `example_codelet.o`
* `schema:manual_ctrl_event_serializer.so` - serializer library for `manual_ctrl_event` protobuf struct.
* `schema:packet_serializer.so` - serializer library for `packet` protobuf struct.
* `schema.pb` - compiled protobuf of [schema.proto](./schema.proto).
* `schema.pb.c` - nanopb generated C file.
* `schema.pb.h` - nanopb generated H file.

To bring the primary application up, we run the following commands:
```sh
$ source ../../setup_jbpfp_env.sh
$ ./run_collect_control.sh
```

To start the local decoder:
```sh
$ source ../../setup_jbpfp_env.sh
$ ./run_decoder.sh
```

If successful, we should see the following line printed:
```
[JBPF_INFO]: Allocated size is 1107296256
```

To bring the primary application up, we run the following commands on a second terminal:
```sh
$ source ../../setup_jbpfp_env.sh
$ ./run_app.sh
```

If successful, we should see the following printed in the log of the secondary:
```
[JBPF_INFO]: Agent thread initialization finished
[JBPF_INFO]: Setting the name of thread 1035986496 to jbpf_lcm_ipc
[JBPF_INFO]: Registered thread id 1
[JBPF_INFO]: Started LCM IPC thread at /var/run/jbpf/jbpf_lcm_ipc
[JBPF_DEBUG]: jbpf_lcm_ipc thread ready
[JBPF_INFO]: Registered thread id 2
[JBPF_INFO]: Started LCM IPC server
```

and on the primary:
```
[JBPF_INFO]: Negotiation was successful
[JBPF_INFO]: Allocation worked for size 1073741824
[JBPF_INFO]: Allocated size is 1073741824
[JBPF_INFO]: Heap was created successfully
```

To load the codeletset, we run the following commands on a third terminal window:
```sh
$ source ../../setup_jbpfp_env.sh
$ ./load.sh
```

If the codeletset was loaded successfully, we should see the following output in the `example_app` window:
```
[JBPF_INFO]: VM created and loaded successfully: example_codelet
```

After that, the primary `example_collect_control` should start printing periodical messages (once per second):
```
INFO[0008] {"seqNo":5, "value":-5, "name":"instance 5"}  streamUUID=00112233-4455-6677-8899-aabbccddeeff
INFO[0009] {"seqNo":6, "value":-6, "name":"instance 6"}  streamUUID=00112233-4455-6677-8899-aabbccddeeff
INFO[0010] {"seqNo":7, "value":-7, "name":"instance 7"}  streamUUID=00112233-4455-6677-8899-aabbccddeeff
```

To send a manual control message to the `example_app`, we run the command:
```sh
$ ./send_input_msg.sh 101
```

This should trigger a message in the `example_app`:
```
[JBPF_DEBUG]:  Called 2 times so far and received manual_ctrl_event with value 101
```

To unload the codeletset, we run the command:
```sh
$ ./unload.sh
```

The `example_app` should stop printing the periodical messages and should give the following output:
```
[JBPF_INFO]: VM with vmfd 0 (i = 0) destroyed successfully
```