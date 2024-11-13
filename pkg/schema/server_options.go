// Copyright (c) Microsoft Corporation. All rights reserved.

package schema

import (
	"errors"
	"jbpf_protobuf_cli/jbpf"

	"github.com/spf13/pflag"
)

// ServerOptions is the options for the decoder server
type ServerOptions struct {
	control *controlOptions
	jbpf    *jbpf.Options
}

// AddServerOptionsToFlags adds the server options to the flags
func AddServerOptionsToFlags(flags *pflag.FlagSet, opts *ServerOptions) {
	if opts == nil {
		return
	}
	if opts.control == nil {
		opts.control = &controlOptions{}
	}
	if opts.jbpf == nil {
		opts.jbpf = &jbpf.Options{}
	}

	addControlOptionsToFlags(flags, opts.control)
	jbpf.AddOptionsToFlags(flags, opts.jbpf)
}

// Parse parses the server options
func (o *ServerOptions) Parse() error {
	return errors.Join(
		o.control.parse(),
		o.jbpf.Parse(),
	)
}
