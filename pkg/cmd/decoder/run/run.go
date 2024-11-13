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
	general *common.GeneralOptions
	data    *data.ServerOptions
	schema  *schema.ServerOptions
}

// Command Run decoder to collect, decode and print jbpf output
func Command(opts *common.GeneralOptions) *cobra.Command {
	runOptions := &runOptions{
		general: opts,
		data:    &data.ServerOptions{},
		schema:  &schema.ServerOptions{},
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
	schema.AddServerOptionsToFlags(cmd.PersistentFlags(), runOptions.schema)
	data.AddServerOptionsToFlags(cmd.PersistentFlags(), runOptions.data)
	return cmd
}

func run(cmd *cobra.Command, opts *runOptions) error {
	if err := errors.Join(
		opts.general.Parse(),
		opts.data.Parse(),
		opts.schema.Parse(),
	); err != nil {
		return err
	}

	logger := opts.general.Logger

	store := schema.NewStore()

	schemaServer, err := schema.NewServer(cmd.Context(), logger, opts.schema, store)
	if err != nil {
		return err
	}

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
