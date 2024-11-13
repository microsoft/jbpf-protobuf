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
	schema  *schema.ClientOptions
	general *common.GeneralOptions

	compiledProtos map[string]*common.File
	configFiles    []string
	configs        []DecoderLoadConfig
}

func addToFlags(flags *pflag.FlagSet, opts *runOptions) {
	flags.StringArrayVarP(&opts.configFiles, "config", "c", []string{}, "configuration files to load")
}

func (o *runOptions) parse() error {
	configs, compiledProtos, err := fromFiles(o.configFiles...)
	if err != nil {
		return err
	}
	o.configs = configs
	o.compiledProtos = compiledProtos

	return nil
}

// Command Load a schema to a local decoder
func Command(opts *common.GeneralOptions) *cobra.Command {
	runOptions := &runOptions{
		schema:  &schema.ClientOptions{},
		general: opts,
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
	schema.AddClientOptionsToFlags(cmd.PersistentFlags(), runOptions.schema)
	return cmd
}

func run(cmd *cobra.Command, opts *runOptions) error {
	if err := errors.Join(
		opts.general.Parse(),
		opts.schema.Parse(),
		opts.parse(),
	); err != nil {
		return err
	}

	logger := opts.general.Logger

	client, err := schema.NewClient(cmd.Context(), logger, opts.schema)
	if err != nil {
		return err
	}

	schemas := make(map[string]*schema.LoadRequest)

	for _, config := range opts.configs {
		for _, desc := range config.CodeletDescriptor {
			for _, io := range desc.InIOChannel {
				if existing, ok := schemas[io.Serde.Protobuf.protoPackageName]; ok {
					existing.Streams[io.streamUUID] = io.Serde.Protobuf.MsgName
				} else {
					compiledProto := opts.compiledProtos[io.Serde.Protobuf.absPackagePath]
					schemas[io.Serde.Protobuf.protoPackageName] = &schema.LoadRequest{
						CompiledProto: compiledProto.Data,
						Streams: map[uuid.UUID]string{
							io.streamUUID: io.Serde.Protobuf.MsgName,
						},
					}
				}
			}
			for _, io := range desc.OutIOChannel {
				if existing, ok := schemas[io.Serde.Protobuf.protoPackageName]; ok {
					existing.Streams[io.streamUUID] = io.Serde.Protobuf.MsgName
				} else {
					compiledProto := opts.compiledProtos[io.Serde.Protobuf.absPackagePath]
					schemas[io.Serde.Protobuf.protoPackageName] = &schema.LoadRequest{
						CompiledProto: compiledProto.Data,
						Streams: map[uuid.UUID]string{
							io.streamUUID: io.Serde.Protobuf.MsgName,
						},
					}
				}
			}
		}
	}

	return client.Load(schemas)
}
