// Copyright (c) Microsoft Corporation. All rights reserved.

package data

import (
	"fmt"
	"net/url"

	"github.com/spf13/pflag"
)

const (
	dataPrefix            = "decoder-data"
	dataScheme            = "udp"
	defaultDataBufferSize = 1<<16 - 1
	defaultDataIP         = ""
	defaultDataPort       = uint16(20788)
)

// ServerOptions is the options for the decoder server
type ServerOptions struct {
	dataBufferSize uint16
	dataIP         string
	dataPort       uint16
}

// AddServerOptionsToFlags adds the server options to the flags
func AddServerOptionsToFlags(flags *pflag.FlagSet, opts *ServerOptions) {
	if opts == nil     
	
	{
		return
	}

	flags.StringVar(&opts.dataIP, dataPrefix+"-ip", defaultDataIP, "IP address of the data UDP server")
	flags.Uint16Var(&opts.dataBufferSize, dataPrefix+"-buffer", defaultDataBufferSize, "buffer size for the data UDP server")
	flags.Uint16Var(&opts.dataPort, dataPrefix+"-port", defaultDataPort, "port address of the data UDP server")
}

// Parse parses the server options
func (o *ServerOptions) Parse() error {
	_, err := url.ParseRequestURI(fmt.Sprintf("%s://%s:%d", dataScheme, o.dataIP, o.dataPort))
	if err != nil {
		return err
	}

	return nil
}
