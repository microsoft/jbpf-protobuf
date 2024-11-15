// Copyright (c) Microsoft Corporation. All rights reserved.

package schema

import (
	"fmt"
	"net/url"

	"github.com/spf13/pflag"
)

const (
	// DefaultControlPort is the default used for the local decoder server
	DefaultControlPort = uint16(20789)

	controlPrefix    = "decoder-api"
	controlScheme    = "http"
	defaultControlIP = ""
)

// Options for internal communication with the decoder
type Options struct {
	ip   string
	port uint16
}

// AddOptionsToFlags adds the options to the provided flag set
func AddOptionsToFlags(flags *pflag.FlagSet, opts *Options) {
	flags.StringVar(&opts.ip, controlPrefix+"-ip", defaultControlIP, "IP address of the decoder HTTP server")
	flags.Uint16Var(&opts.port, controlPrefix+"-port", DefaultControlPort, "port address of the decoder HTTP server")
}

// Parse the options
func (o *Options) Parse() error {
	_, err := url.ParseRequestURI(fmt.Sprintf("%s://%s:%d", controlScheme, o.ip, o.port))
	if err != nil {
		return err
	}

	return nil
}
