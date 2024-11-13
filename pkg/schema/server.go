// Copyright (c) Microsoft Corporation. All rights reserved.

package schema

import (
	context "context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"jbpf_protobuf_cli/jbpf"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Server is a server that implements the DynamicDecoderServer interface
type Server struct {
	ctx        context.Context
	jbpfClient *jbpf.Client
	logger     *logrus.Logger
	opts       *ServerOptions
	store      *Store
}

// NewServer returns a new Server
func NewServer(ctx context.Context, logger *logrus.Logger, opts *ServerOptions, store *Store) (*Server, error) {
	var jbpfClient *jbpf.Client
	var err error

	if opts.jbpf.Enable {
		jbpfClient, err = jbpf.NewClient(logger, opts.jbpf)
		if err != nil {
			return nil, err
		}
	}

	return &Server{
		ctx:        ctx,
		jbpfClient: jbpfClient,
		logger:     logger,
		opts:       opts,
		store:      store,
	}, nil
}

// Serve starts the server
func (s *Server) Serve() error {
	return s.serveHTTP(s.ctx)
}

// UpsertProtoPackage registers a proto package with the server
func (s *Server) UpsertProtoPackage(_ context.Context, req *UpsertSchemaRequest) error {
	checksum := sha1.Sum(req.ProtoDescriptor)
	checksumAsString := base64.StdEncoding.EncodeToString(checksum[:])

	fds := &descriptorpb.FileDescriptorSet{}
	if err := proto.Unmarshal(req.ProtoDescriptor, fds); err != nil {
		s.logger.WithError(err).Error("unable to unmarshal proto descriptor")
		return err
	}

	protoPackageFile := fds.File[0].GetName()
	protoPackageName := strings.TrimSuffix(protoPackageFile, filepath.Ext(protoPackageFile))
	l := s.logger.WithFields(logrus.Fields{
		"protoPackageName": protoPackageName,
		"checksum":         checksumAsString,
	})

	if len(fds.File) != 1 {
		err := fmt.Errorf("expected exactly one file descriptor in the set, got %d", len(fds.File))
		l.WithError(err).Error("unable to interpret proto descriptor")
		return err
	}

	if current, ok := s.store.schemas[protoPackageName]; ok {
		if current.checksum == checksum {
			l.Info("checksum matches, skipping")
			return nil
		}
		l.Warn("overwriting existing proto package")
	} else {
		l.Info("setting proto package")
	}

	s.store.schemas[protoPackageName] = &RecordedProtoDescriptor{
		checksum:        checksum,
		ProtoDescriptor: req.ProtoDescriptor,
	}

	return nil
}

// AddStreamToSchemaAssociation associates a stream with a schema
func (s *Server) AddStreamToSchemaAssociation(_ context.Context, req *AddSchemaAssociationRequest) error {
	l := s.logger.WithFields(logrus.Fields{
		"protoMsg":     req.ProtoMessage,
		"protoPackage": req.ProtoPackage,
		"streamUUID":   req.StreamUUID.String(),
	})

	if current, ok := s.store.streamToSchema[req.StreamUUID]; ok {
		if current.ProtoMsg == req.ProtoMessage && current.ProtoPackage == req.ProtoPackage {
			return nil
		}
		err := fmt.Errorf("stream already has a schema association")
		l.WithError(err).Error("error adding stream to schema association")
		return err
	}

	if _, ok := s.store.schemas[req.ProtoPackage]; !ok {
		err := fmt.Errorf("proto package %s not found", req.ProtoPackage)
		l.WithError(err).Error("error adding stream to schema association")
		return err
	}

	s.store.streamToSchema[req.StreamUUID] = &RecordedStreamToSchema{
		ProtoMsg:     req.ProtoMessage,
		ProtoPackage: req.ProtoPackage,
	}

	l.Info("association added")

	return nil
}

// SendControl sends data to the jbpf agent
func (s *Server) SendControl(_ context.Context, req *SendControlRequest) error {
	msg, err := s.store.GetProtoMsgInstance(req.StreamUUID)
	if err != nil {
		s.logger.WithError(err).Errorf("error creating instance of proto message %s", req.StreamUUID.String())
		return err
	}

	err = protojson.Unmarshal([]byte(req.Payload), msg)
	if err != nil {
		s.logger.WithError(err).Error("error unmarshalling payload")
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"msg": fmt.Sprintf("%T - \"%v\"", msg, msg),
	}).Info("sending msg")

	payload, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	out := append(req.StreamUUID[:], payload...)
	if err := s.jbpfClient.Write(out); err != nil {
		return err
	}

	return nil
}

// DeleteStreamToSchemaAssociation removes the association between a stream and a schema
func (s *Server) DeleteStreamToSchemaAssociation(_ context.Context, req uuid.UUID) error {
	l := s.logger.WithField("streamUUID", req.String())

	if current, ok := s.store.streamToSchema[req]; !ok {
		l.Debug("no association found for stream UUID")
	} else {
		delete(s.store.streamToSchema, req)
		l.WithFields(logrus.Fields{
			"protoMsg":     current.ProtoMsg,
			"protoPackage": current.ProtoPackage,
		}).Info("association removed")
	}

	return nil
}
