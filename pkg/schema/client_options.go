// Copyright (c) Microsoft Corporation. All rights reserved.

package schema

import "github.com/spf13/pflag"

// ClientOptions is the options for the decoder client
type ClientOptions struct {
	control *controlOptions
}

// AddClientOptionsToFlags adds the client options to the flags
func AddClientOptionsToFlags(flags *pflag.FlagSet, opts *ClientOptions) {
	if opts.control == nil {
		opts.control = &controlOptions{}
	}

	addControlOptionsToFlags(flags, opts.control)
}

// Parse parses the client options
func (o *ClientOptions) Parse() error {
	return o.control.parse()
}
