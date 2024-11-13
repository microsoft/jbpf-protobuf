#define PB_FIELD_32BIT 1
#include <pb.h>
#include <pb_decode.h>
#include <pb_encode.h>
#include "example.pb.h"

const uint32_t proto_message_size = sizeof(req_resp);

int jbpf_io_serialize(void* input_msg_buf, size_t input_msg_buf_size, char* serialized_data_buf, size_t serialized_data_buf_size) {
	if (input_msg_buf_size != proto_message_size)
		return -1;

	pb_ostream_t ostream = pb_ostream_from_buffer((uint8_t*)serialized_data_buf, serialized_data_buf_size);
	if (!pb_encode(&ostream, req_resp_fields, input_msg_buf))
		return -1;

	return ostream.bytes_written;
}

int jbpf_io_deserialize(char* serialized_data_buf, size_t serialized_data_buf_size, void* output_msg_buf, size_t output_msg_buf_size) {
	if (output_msg_buf_size != proto_message_size)
		return 0;

	pb_istream_t istream = pb_istream_from_buffer((uint8_t*)serialized_data_buf, serialized_data_buf_size);
	return pb_decode(&istream, req_resp_fields, output_msg_buf);
}
