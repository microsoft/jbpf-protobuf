// Copyright (c) Microsoft Corporation. All rights reserved.

package unload

import (
	"errors"
	"jbpf_protobuf_cli/common"
	"jbpf_protobuf_cli/schema"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	maxStreamUUIDs = 1024
)

type runOptions struct {
	decoderAPI *schema.Options
	general    *common.GeneralOptions

	configFiles []string
	configs     []common.CodeletsetConfig
}

func addToFlags(flags *pflag.FlagSet, opts *runOptions) {
	flags.StringArrayVarP(&opts.configFiles, "config", "c", []string{}, "configuration files to unload")
}

func (o *runOptions) parse() error {
	configs, _, err := common.CodeletsetConfigFromFiles(o.configFiles...)
	if err != nil {
		return err
	}
	o.configs = configs

	return nil
}

// Command Unload a schema from a local decoder
func Command(opts *common.GeneralOptions) *cobra.Command {
	runOptions := &runOptions{
		decoderAPI: &schema.Options{},
		general:    opts,
	}
	cmd := &cobra.Command{
		Use:   "unload",
		Short: "Unload a schema from a local decoder",
		Long:  "Unload a schema from a local decoder",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd, runOptions)
		},
		SilenceUsage: true,
	}
	addToFlags(cmd.PersistentFlags(), runOptions)
	schema.AddOptionsToFlags(cmd.PersistentFlags(), runOptions.decoderAPI)
	return cmd
}

func run(cmd *cobra.Command, opts *runOptions) error {
	if err := errors.Join(
		opts.general.Parse(),
		opts.decoderAPI.Parse(),
		opts.parse(),
	); err != nil {
		return err
	}

	logger := opts.general.Logger

	client, err := schema.NewClient(cmd.Context(), logger, opts.decoderAPI)
	if err != nil {
		return err
	}

	streamUUIDs := make([]uuid.UUID, 0, maxStreamUUIDs)

	for _, config := range opts.configs {
		for _, desc := range config.CodeletDescriptor {
			for _, io := range desc.OutIOChannel {
				streamUUIDs = append(streamUUIDs, io.StreamUUID)
			}
		}
	}

	return client.Unload(streamUUIDs)
}
