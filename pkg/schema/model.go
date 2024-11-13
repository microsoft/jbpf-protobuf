// Copyright (c) Microsoft Corporation. All rights reserved.

package schema

import (
	"encoding/base64"
	"encoding/json"

	"github.com/google/uuid"
)

// UpsertSchemaRequest is the request body for the /schema endpoint
type UpsertSchemaRequest struct {
	ProtoDescriptor []byte
}

// MarshalJSON marshals the UpsertSchemaRequest to JSON
func (u UpsertSchemaRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ProtoDescriptor string
	}{
		ProtoDescriptor: base64.StdEncoding.EncodeToString(u.ProtoDescriptor),
	})
}

// UnmarshalJSON unmarshals the UpsertSchemaRequest from JSON
func (u *UpsertSchemaRequest) UnmarshalJSON(data []byte) error {
	var intermediate struct{ ProtoDescriptor string }
	if err := json.Unmarshal(data, &intermediate); err != nil {
		return err
	}
	protoDesc, err := base64.StdEncoding.DecodeString(intermediate.ProtoDescriptor)
	if err != nil {
		return err
	}
	u.ProtoDescriptor = protoDesc
	return nil
}

// AddSchemaAssociationRequest is the request body for the /stream endpoint
type AddSchemaAssociationRequest struct {
	StreamUUID   uuid.UUID
	ProtoPackage string
	ProtoMessage string
}

// MarshalJSON marshals the AddSchemaAssociationRequest to JSON
func (a AddSchemaAssociationRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		StreamUUID   string
		ProtoPackage string
		ProtoMessage string
	}{
		StreamUUID:   a.StreamUUID.String(),
		ProtoPackage: a.ProtoPackage,
		ProtoMessage: a.ProtoMessage,
	})
}

// UnmarshalJSON unmarshals the AddSchemaAssociationRequest from JSON
func (a *AddSchemaAssociationRequest) UnmarshalJSON(data []byte) error {
	var intermediate struct {
		StreamUUID   string
		ProtoPackage string
		ProtoMessage string
	}
	if err := json.Unmarshal(data, &intermediate); err != nil {
		return err
	}
	streamUUID, err := uuid.Parse(intermediate.StreamUUID)
	if err != nil {
		return err
	}
	a.StreamUUID = streamUUID
	a.ProtoPackage = intermediate.ProtoPackage
	a.ProtoMessage = intermediate.ProtoMessage
	return nil
}

// SendControlRequest is the request body for the /control endpoint
type SendControlRequest struct {
	StreamUUID uuid.UUID
	Payload    string
}

// MarshalJSON marshals the SendControlRequest to JSON
func (s SendControlRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		StreamUUID string
		Payload    string
	}{
		StreamUUID: s.StreamUUID.String(),
		Payload:    s.Payload,
	})
}

// UnmarshalJSON unmarshals the SendControlRequest from JSON
func (s *SendControlRequest) UnmarshalJSON(data []byte) error {
	var intermediate struct {
		StreamUUID string
		Payload    string
	}
	if err := json.Unmarshal(data, &intermediate); err != nil {
		return err
	}
	streamUUID, err := uuid.Parse(intermediate.StreamUUID)
	if err != nil {
		return err
	}
	s.StreamUUID = streamUUID
	s.Payload = intermediate.Payload
	return nil
}
