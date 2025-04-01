// Copyright (c) Microsoft Corporation. All rights reserved.

package schema

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	defaultIPAddr = "localhost"
)

// Client encapsulates the decoder client
type Client struct {
	baseURL string
	ctx     context.Context
	inner   *http.Client
	logger  *logrus.Logger
}

// NewClient creates a new Client
func NewClient(ctx context.Context, logger *logrus.Logger, opts *Options) (*Client, error) {
	ip := opts.ip
	if len(ip) == 0 {
		ip = defaultIPAddr
	}

	return &Client{
		baseURL: fmt.Sprintf("%s://%s:%d", controlScheme, ip, opts.port),
		ctx:     ctx,
		inner:   &http.Client{},
		logger:  logger,
	}, nil
}

func (c *Client) doPost(relativePath string, input interface{}) error {
	jsonData, err := json.Marshal(input)
	if err != nil {
		return err
	}

	var req *http.Request
	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", c.baseURL, relativePath), bytes.NewReader(jsonData))
	if err != nil {
		return err
	}

	resp, err := c.inner.Do(req)
	if err != nil {
		c.logger.WithError(err).Error("http request failed")
		return err
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		err := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		c.logger.WithField("body", buf.String()).WithError(err).Error("unexpected status code")
		return err
	}

	return nil
}

func (c *Client) doDelete(relativePath string) error {
	var req *http.Request
	var err error
	req, err = http.NewRequest(http.MethodDelete, fmt.Sprintf("%s%s", c.baseURL, relativePath), nil)
	if err != nil {
		return err
	}

	resp, err := c.inner.Do(req)
	if err != nil {
		c.logger.WithError(err).Error("http request failed")
		return err
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		err := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		c.logger.WithField("body", buf.String()).WithError(err).Error("unexpected status code")
		return err
	}

	return nil
}

// LoadRequest is a request to load a schema and stream
type LoadRequest struct {
	CompiledProto []byte
	Streams       map[uuid.UUID]string
}

// Load loads the schemas into the decoder
func (c *Client) Load(schemas map[string]*LoadRequest) error {
	errs := make([]error, 0, len(schemas))

	for protoPackageName, req := range schemas {
		l := c.logger.WithFields(logrus.Fields{"pkg": protoPackageName})

		if err := c.doPost("/schema", &UpsertSchemaRequest{ProtoDescriptor: req.CompiledProto}); err != nil {
			err = fmt.Errorf("failed to upsert proto package %s: %w", protoPackageName, err)
			errs = append(errs, err)
			continue
		}

		l.Info("successfully upserted proto package")

		for streamUUID, protoMsg := range req.Streams {
			err := c.doPost("/stream", &AddSchemaAssociationRequest{StreamUUID: streamUUID, ProtoPackage: protoPackageName, ProtoMessage: protoMsg})
			if err != nil {
				err = fmt.Errorf("failed to associate streamID %s to proto package %s and message %s: %w", streamUUID.String(), protoPackageName, protoMsg, err)
				errs = append(errs, err)
				continue
			}

			l.WithFields(logrus.Fields{
				"protoMsg":         protoMsg,
				"protoPackageName": protoPackageName,
				"streamId":         streamUUID.String(),
			}).Info("successfully associated stream ID with proto package")
		}
	}

	return errors.Join(errs...)
}

// SendControl dispatches a control message to the decoder
func (c *Client) SendControl(streamUUID uuid.UUID, jdata string) error {
	if err := c.doPost("/control", &SendControlRequest{StreamUUID: streamUUID, Payload: jdata}); err != nil {
		return fmt.Errorf("failed to send control message %s: %w", streamUUID.String(), err)
	}

	c.logger.WithFields(logrus.Fields{
		"streamId": streamUUID.String(),
	}).Info("successfully sent control message")

	return nil
}

// Unload removes the stream association from the decoder
func (c *Client) Unload(streamUUIDs []uuid.UUID) error {
	errs := make([]error, 0, len(streamUUIDs))
	for _, streamUUID := range streamUUIDs {
		// using base64.RawURLEncoding to encode the streamUUID to a URL-safe string
		streamIDStr := base64.RawURLEncoding.EncodeToString(streamUUID[:])
		if err := c.doDelete(fmt.Sprintf("/stream?stream_uuid=%s", streamIDStr)); err != nil {
			err = fmt.Errorf("failed to delete stream ID association %s: %w", streamUUID.String(), err)
			errs = append(errs, err)
			continue
		}

		c.logger.WithFields(logrus.Fields{
			"streamId": streamUUID.String(),
		}).Info("successfully deleted stream ID association")
	}

	return errors.Join(errs...)
}
