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

	controlPrefix    = "decoder-control"
	controlScheme    = "http"
	defaultControlIP = ""
)

type controlOptions struct {
	ip   string
	port uint16
}

func addControlOptionsToFlags(flags *pflag.FlagSet, opts *controlOptions) {
	flags.StringVar(&opts.ip, controlPrefix+"-ip", defaultControlIP, "IP address of the control HTTP server")
	flags.Uint16Var(&opts.port, controlPrefix+"-port", DefaultControlPort, "port address of the control HTTP server")
}

func (o *controlOptions) parse() error {
	_, err := url.ParseRequestURI(fmt.Sprintf("%s://%s:%d", controlScheme, o.ip, o.port))
	if err != nil {
		return err
	}

	return nil
}
