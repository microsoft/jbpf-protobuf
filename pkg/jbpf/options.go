// Copyright (c) Microsoft Corporation. All rights reserved.

package jbpf

import (
	"fmt"
	"net/url"
	"time"

	"github.com/spf13/pflag"
)

const (
	defaultIP     = ""
	defaultPort   = uint16(20787)
	optionsPrefix = "jbpf"
	scheme        = "tcp"
)

// Options is the options for the jbpf client
type Options struct {
	ip              string
	keepAlivePeriod time.Duration
	port            uint16
}

// AddOptionsToFlags adds the options to the flags
func AddOptionsToFlags(flags *pflag.FlagSet, opts *Options) {
	if opts == nil {
		return
	}

	flags.DurationVar(&opts.keepAlivePeriod, optionsPrefix+"-keep-alive", 0, "time to keep alive the connection")
	flags.StringVar(&opts.ip, optionsPrefix+"-ip", defaultIP, "IP address of the jbpf TCP server")
	flags.Uint16Var(&opts.port, optionsPrefix+"-port", defaultPort, "port address of the jbpf TCP server")
}

// Parse parses the options
func (o *Options) Parse() error {
	_, err := url.ParseRequestURI(fmt.Sprintf("%s://%s:%d", scheme, o.ip, o.port))
	if err != nil {
		return err
	}

	return nil
}
