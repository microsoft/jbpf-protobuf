// Copyright (c) Microsoft Corporation. All rights reserved.

package data

import (
	context "context"
	"errors"
	"fmt"
	"jbpf_protobuf_cli/schema"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	dataReadDeadline = 1 * time.Second
	decoderChanSize  = 100
)

// Server is a server that implements the DynamicDecoderServer interface
type Server struct {
	ctx    context.Context
	logger *logrus.Logger
	opts   *ServerOptions
	store  *schema.Store
}

// NewServer returns a new Server
func NewServer(ctx context.Context, logger *logrus.Logger, opts *ServerOptions, store *schema.Store) (*Server, error) {
	return &Server{
		ctx:    ctx,
		logger: logger,
		opts:   opts,
		store:  store,
	}, nil
}

// Listen starts the server
func (s *Server) Listen(onData func(uuid.UUID, []byte)) error {
	data, err := net.ListenPacket(dataScheme, fmt.Sprintf("%s:%d", s.opts.dataIP, s.opts.dataPort))
	if err != nil {
		return err
	}
	s.logger.WithField("addr", data.LocalAddr().Network()+"://"+data.LocalAddr().String()).Debug("starting data server")
	defer func() {
		s.logger.WithField("addr", data.LocalAddr().Network()+"://"+data.LocalAddr().String()).Debug("stopping data server")
		if err := data.Close(); err != nil {
			s.logger.WithError(err).Errorf("error closing data server")
		}
	}()

	stopper := make(chan os.Signal, 1)
	signal.Notify(stopper, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case <-stopper:
			return nil
		case <-s.ctx.Done():
			return nil

		default:
			buffer := make([]byte, s.opts.dataBufferSize)
			if err := data.SetReadDeadline(time.Now().Add(dataReadDeadline)); err != nil {
				return err
			}
			n, _, err := data.ReadFrom(buffer)
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if err != nil {
				return errors.Join(err, fmt.Errorf("error reading from UDP socket"))
			}

			if n < 16 {
				s.logger.Warnf("received data is less than %d bytes, skipping", 16)
				continue
			}

			streamUUID, err := uuid.FromBytes(buffer[:16])
			if err != nil {
				s.logger.WithError(err).Error("error parsing stream UUID")
				continue
			}

			msg, err := s.store.GetProtoMsgInstance(streamUUID)
			if err != nil {
				s.logger.WithError(err).Error("error creating instance of proto message")
				continue
			}

			err = proto.Unmarshal(buffer[16:n], msg)
			if err != nil {
				s.logger.WithError(err).Error("error unmarshalling payload")
				continue
			}

			res, err := protojson.Marshal(msg)
			if err != nil {
				s.logger.WithError(err).Error("error marshalling message to JSON")
				continue
			}

			onData(streamUUID, res)
		}
	}

}
