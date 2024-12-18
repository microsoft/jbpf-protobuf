# Basic example of standalone *jbpf* operation

This example showcases a basic *jbpf* usage scenario. It provides a dummy C++ application (`example_app`), that initializes
*jbpf* in standalone mode, and an example codelet (`example_codelet.o`).
The example demonstrates the following:
1. How to declare and call hooks.
2. How to register handler functions for capturing output data from codelets in standalone mode.
3. How to load and unload codeletsets using the LCM CLI tool (via a Unix socket API).
4. How to send data back to running codelets.

For more details of the exact behavior of the application and the codelet, check [here](../../docs/understand_first_codelet.md).
You can also check the inline comments in [example_app.cpp](./example_app.cpp)
and [example_codelet.c](./example_codelet.c)

## Usage

Before running the sample apps build the project, see [README.md](../../README.md) for instructions to build bare metal or with containers.

### Running bare metal

To build the example from scratch, we run the following commands:
```sh
$ source ../../setup_jbpfp_env.sh
$ make
```

This should produce these artifacts:
* `example_app`
* `example_codelet.o`
* `schema:manual_ctrl_event_serializer.so` - serializer library for `manual_ctrl_event` protobuf struct.
* `schema:packet_serializer.so` - serializer library for `packet` protobuf struct.
* `schema.pb` - compiled protobuf of [schema.proto](./schema.proto).
* `schema.pb.c` - nanopb generated C file.
* `schema.pb.h` - nanopb generated H file.

To bring up the application, we run the following commands:
```sh
$ source ../../setup_jbpfp_env.sh
$ ./run_app.sh
```

To start the local decoder:
```sh
$ source ../../setup_jbpfp_env.sh
$ ./run_decoder.sh
```

If successful, we shoud see the following line printed:
```
[JBPF_INFO]: Started LCM IPC server
```

To load the codeletset, we run the following commands on a second terminal window:
```sh
$ source ../../setup_jbpfp_env.sh
$ ./load.sh
```

If the codeletset was loaded successfully, we should see the following output in the `example_app` window:
```
[JBPF_INFO]: VM created and loaded successfully: example_codelet
```

After that, the agent should start printing periodical messages (once per second):
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

### Running with containers

To build the example from scratch, we run the following commands:
```sh
$ docker run --rm -it \
  -v $(dirname $(dirname $(pwd))):/jbpf-protobuf \
  -w /jbpf-protobuf/examples/first_example_standalone \
    jbpfp-$OS:latest make
```

This should produce these artifacts:
* `example_app`
* `example_codelet.o`
* `schema:manual_ctrl_event_serializer.so` - serializer library for `manual_ctrl_event` protobuf struct.
* `schema:packet_serializer.so` - serializer library for `packet` protobuf struct.
* `schema.pb` - compiled protobuf of [schema.proto](./schema.proto).
* `schema.pb.c` - nanopb generated C file.
* `schema.pb.h` - nanopb generated H file.

To bring up the application, we run the following commands:
```sh
$ docker run --rm -it --net=host \
  -v /tmp/jbpf:/tmp/jbpf \
  -v /dev/shm:/dev/shm \
  -v $(dirname $(dirname $(pwd))):/jbpf-protobuf \
  -w /jbpf-protobuf/examples/first_example_standalone \
    jbpfp-$OS:latest ./run_app.sh
```

To start the local decoder:
```sh
$ docker run --rm -it --net=host \
  -v $(dirname $(dirname $(pwd))):/jbpf-protobuf \
  -w /jbpf-protobuf/examples/first_example_standalone \
    jbpfp-$OS:latest ./run_decoder.sh
```

If successful, we shoud see the following line printed:
```
[JBPF_INFO]: Started LCM IPC server
```

To load the codeletset, we run the following commands on a second terminal window:
```sh
$ docker run --rm -it --net=host \
  -v /tmp/jbpf:/tmp/jbpf \
  -v $(dirname $(dirname $(pwd))):/jbpf-protobuf \
  -w /jbpf-protobuf/examples/first_example_standalone \
    jbpfp-$OS:latest ./load.sh
```

If the codeletset was loaded successfully, we should see the following output in the `example_app` window:
```
[JBPF_INFO]: VM created and loaded successfully: example_codelet
```

After that, the agent should start printing periodical messages (once per second):
```
INFO[0008] {"seqNo":5, "value":-5, "name":"instance 5"}  streamUUID=00112233-4455-6677-8899-aabbccddeeff
INFO[0009] {"seqNo":6, "value":-6, "name":"instance 6"}  streamUUID=00112233-4455-6677-8899-aabbccddeeff
INFO[0010] {"seqNo":7, "value":-7, "name":"instance 7"}  streamUUID=00112233-4455-6677-8899-aabbccddeeff
```

To send a manual control message to the `example_app`, we run the command:
```sh
$ docker run --rm -it --net=host \
  -v /tmp/jbpf:/tmp/jbpf \
  -v $(dirname $(dirname $(pwd))):/jbpf-protobuf \
  -w /jbpf-protobuf/examples/first_example_standalone \
    jbpfp-$OS:latest ./send_input_msg.sh 101
```

This should trigger a message in the `example_app`:
```
[JBPF_DEBUG]:  Called 2 times so far and received manual_ctrl_event with value 101
```

To unload the codeletset, we run the command:
```sh
$ docker run --rm -it --net=host \
  -v /tmp/jbpf:/tmp/jbpf \
  -v $(dirname $(dirname $(pwd))):/jbpf-protobuf \
  -w /jbpf-protobuf/examples/first_example_standalone \
    jbpfp-$OS:latest ./unload.sh
```

The `example_app` should stop printing the periodical messages and should give the following output:
```
[JBPF_INFO]: VM with vmfd 0 (i = 0) destroyed successfully
```