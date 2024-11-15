// Copyright (c) Microsoft Corporation. All rights reserved.

package common

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// NewGeneralOptions creates a new GeneralOptions with default values
func NewGeneralOptions(flags *pflag.FlagSet) *GeneralOptions {
	opts := &GeneralOptions{}
	opts.addToFlags(flags)
	return opts
}

// NewGeneralOptionsFromLogger creates a new GeneralOptions from a logger
func NewGeneralOptionsFromLogger(logger *logrus.Logger) *GeneralOptions {
	opts := &GeneralOptions{
		file:         "",
		formatter:    "TextFormatter",
		Logger:       logger,
		logLevel:     logger.Level.String(),
		reportCaller: logger.ReportCaller,
	}
	return opts
}

// GeneralOptions contains the general options for the jbpf cli
type GeneralOptions struct {
	file         string
	formatter    string
	logLevel     string
	reportCaller bool

	Logger *logrus.Logger
}

func (opts *GeneralOptions) addToFlags(flags *pflag.FlagSet) {
	flags.BoolVar(&opts.reportCaller, "log-report-caller", false, "show report caller in logs")
	flags.StringVar(&opts.file, "log-file", "", "if set, will write logs to file as well as terminal")
	flags.StringVar(&opts.formatter, "log-formatter", "TextFormatter", "logger formatter, set to UncoloredTextFormatter, JSONFormatter or TextFormatter")
	flags.StringVar(&opts.logLevel, "log-level", "info", "log level, set to: panic, fatal, error, warn, info, debug or trace")
}

// Parse will process and validate args
func (opts *GeneralOptions) Parse() error {
	var err1, err2 error
	opts.Logger, err1 = opts.getLogger()
	return errors.Join(err1, err2)
}

// GetLogger returns a logger based on the options
func (opts *GeneralOptions) getLogger() (*logrus.Logger, error) {
	logLev, err := logrus.ParseLevel(opts.logLevel)
	if err != nil {
		return nil, err
	}

	var formatter logrus.Formatter
	switch strings.ToLower(opts.formatter) {
	case "uncoloredtextformatter":
		formatter = new(UncoloredTextFormatter)
	case "jsonformatter":
		formatter = new(logrus.JSONFormatter)
	case "textformatter":
		formatter = new(logrus.TextFormatter)
	default:
		return nil, fmt.Errorf("invalid log formatter: %v", opts.formatter)
	}

	var out io.Writer = os.Stderr

	if opts.file != "" {
		file, err := os.OpenFile(opts.file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		out = io.MultiWriter(os.Stderr, file)
	}

	return &logrus.Logger{
		Out:          out,
		Formatter:    formatter,
		Hooks:        make(logrus.LevelHooks),
		Level:        logLev,
		ExitFunc:     os.Exit,
		ReportCaller: opts.reportCaller,
	}, nil
}
