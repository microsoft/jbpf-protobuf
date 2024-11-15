// Copyright (c) Microsoft Corporation. All rights reserved.

package run

import (
	"errors"
	"jbpf_protobuf_cli/common"
	"jbpf_protobuf_cli/data"
	"jbpf_protobuf_cli/schema"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

type runOptions struct {
	general    *common.GeneralOptions
	data       *data.ServerOptions
	decoderAPI *schema.Options
}

// Command Run decoder to collect, decode and print jbpf output
func Command(opts *common.GeneralOptions) *cobra.Command {
	runOptions := &runOptions{
		general:    opts,
		data:       &data.ServerOptions{},
		decoderAPI: &schema.Options{},
	}
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run decoder to collect, decode and print jbpf output",
		Long:  "Run dynamic protobuf decoder to collect, decode and print jbpf output.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd, runOptions)
		},
		SilenceUsage: true,
	}
	schema.AddOptionsToFlags(cmd.PersistentFlags(), runOptions.decoderAPI)
	data.AddServerOptionsToFlags(cmd.PersistentFlags(), runOptions.data)
	return cmd
}

func run(cmd *cobra.Command, opts *runOptions) error {
	if err := errors.Join(
		opts.general.Parse(),
		opts.data.Parse(),
		opts.decoderAPI.Parse(),
	); err != nil {
		return err
	}

	logger := opts.general.Logger

	store := schema.NewStore()

	schemaServer := schema.NewServer(cmd.Context(), logger, opts.decoderAPI, store)

	dataServer, err := data.NewServer(cmd.Context(), logger, opts.data, store)
	if err != nil {
		return err
	}

	g, _ := errgroup.WithContext(cmd.Context())

	g.Go(func() error {
		return dataServer.Listen(func(streamUUID uuid.UUID, data []byte) {
			logger.WithFields(logrus.Fields{
				"streamUUID": streamUUID.String(),
			}).Info(string(data))
		})
	})

	g.Go(func() error {
		return schemaServer.Serve()
	})

	return g.Wait()
}
