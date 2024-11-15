package schema

import (
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

// RecordedProtoDescriptor is a recorded proto descriptor
type RecordedProtoDescriptor struct {
	checksum        [20]byte
	ProtoDescriptor []byte
}

// RecordedStreamToSchema is a mapping of a stream to a schema
type RecordedStreamToSchema struct {
	ProtoMsg     string
	ProtoPackage string
}

// Store is an in memory store for protobuf schemas
type Store struct {
	schemas        map[string]*RecordedProtoDescriptor
	streamToSchema map[uuid.UUID]*RecordedStreamToSchema
}

// NewStore returns a new Store
func NewStore() *Store {
	return &Store{
		schemas:        make(map[string]*RecordedProtoDescriptor),
		streamToSchema: make(map[uuid.UUID]*RecordedStreamToSchema),
	}
}

// GetProtoMsgInstance returns a new dynamic protobuf message instance
func (s *Store) GetProtoMsgInstance(streamUUID uuid.UUID) (*dynamicpb.Message, error) {
	schema, ok := s.streamToSchema[streamUUID]
	if !ok {
		return nil, fmt.Errorf("no schema found for stream UUID %s", streamUUID.String())
	}

	sch, ok := s.schemas[schema.ProtoPackage]
	if !ok {
		return nil, fmt.Errorf("no schema found for proto package %s", schema.ProtoPackage)
	}

	fds := &descriptorpb.FileDescriptorSet{}
	if err := proto.Unmarshal(sch.ProtoDescriptor, fds); err != nil {
		return nil, err
	}

	pd, err := protodesc.NewFiles(fds)
	if err != nil {
		return nil, err
	}

	msgName := protoreflect.FullName(schema.ProtoMsg)
	var desc protoreflect.Descriptor
	desc, err = pd.FindDescriptorByName(msgName)
	if err != nil {
		return nil, err
	}

	md, ok := desc.(protoreflect.MessageDescriptor)
	if !ok {
		return nil, fmt.Errorf("failed to cast desc to protoreflect.MessageDescriptor, got %T", desc)
	}

	return dynamicpb.NewMessage(md), nil
}
