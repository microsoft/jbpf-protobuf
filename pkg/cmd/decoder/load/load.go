// Copyright (c) Microsoft Corporation. All rights reserved.

package load

import (
	"errors"
	"jbpf_protobuf_cli/common"
	"jbpf_protobuf_cli/schema"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type runOptions struct {
	decoderAPI *schema.Options
	general    *common.GeneralOptions

	compiledProtos map[string]*common.File
	configFiles    []string
	configs        []*common.CodeletsetConfig
}

func addToFlags(flags *pflag.FlagSet, opts *runOptions) {
	flags.StringArrayVarP(&opts.configFiles, "config", "c", []string{}, "configuration files to load")
}

func (o *runOptions) parse() (err error) {
	o.configs, err = common.CodeletsetConfigFromFiles(o.configFiles...)
	if err != nil {
		return
	}
	o.compiledProtos, err = common.LoadCompiledProtos(o.configs, false, true)
	return
}

// Command Load a schema to a local decoder
func Command(opts *common.GeneralOptions) *cobra.Command {
	runOptions := &runOptions{
		decoderAPI: &schema.Options{},
		general:    opts,
	}
	cmd := &cobra.Command{
		Use:   "load",
		Short: "Load a schema to a local decoder",
		Long:  "Load a schema to a local decoder",
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

	schemas := make(map[string]*schema.LoadRequest)

	for _, config := range opts.configs {
		for _, desc := range config.CodeletDescriptor {
			for _, io := range desc.OutIOChannel {
				if existing, ok := schemas[io.Serde.Protobuf.PackageName]; ok {
					existing.Streams[io.StreamUUID] = io.Serde.Protobuf.MsgName
				} else {
					compiledProto, ok := opts.compiledProtos[io.Serde.Protobuf.PackagePath]
					if !ok {
						return errors.New("compiled proto not found")
					}
					schemas[io.Serde.Protobuf.PackageName] = &schema.LoadRequest{
						CompiledProto: compiledProto.Data,
						Streams: map[uuid.UUID]string{
							io.StreamUUID: io.Serde.Protobuf.MsgName,
						},
					}
				}
			}
		}
	}

	return client.Load(schemas)
}
