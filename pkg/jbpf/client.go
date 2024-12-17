// Copyright (c) Microsoft Corporation. All rights reserved.

package jbpf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
)

// Client is a TCP socket client
type Client struct {
	conn   *net.TCPConn
	logger *logrus.Logger
	opts   *Options
}

// NewClient creates a new socket client
func NewClient(logger *logrus.Logger, opts *Options) (*Client, error) {
	c := &Client{
		logger: logger,
		opts:   opts,
	}
	if err := c.connect(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) connect() error {
	ip := c.opts.ip
	if len(ip) == 0 {
		ip = "localhost"
	}

	conn, err := net.Dial(scheme, fmt.Sprintf("%s:%d", ip, c.opts.port))
	if err != nil {
		return err
	}

	tcpc, ok := conn.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("expected a tcp connection")
	}

	if c.opts.keepAlivePeriod != 0 {
		if err := tcpc.SetKeepAlive(true); err != nil {
			return err
		}
		if err := tcpc.SetKeepAlivePeriod(c.opts.keepAlivePeriod); err != nil {
			return err
		}
	}

	c.conn = tcpc
	return nil
}

// Write writes data to the socket
func (c *Client) Write(bs []byte) error {
	if c.conn == nil {
		if err := c.connect(); err != nil {
			return err
		}
	}

	lengthField := make([]byte, 2)
	binary.LittleEndian.PutUint16(lengthField, uint16(len(bs)))

	if _, err := c.conn.Write(append(lengthField, bs...)); err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) {
			if err := c.Close(); err != nil {
				c.logger.WithError(err).Error("failed to close connection")
			}
			c.conn = nil
			return fmt.Errorf("closing connection: %w", netErr)
		}
		return err
	}

	return nil
}

// Close closes the connection
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	err := c.conn.Close()
	c.conn = nil
	return err
}
