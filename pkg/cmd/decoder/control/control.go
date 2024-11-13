// Copyright (c) Microsoft Corporation. All rights reserved.

package control

import (
	"encoding/json"
	"errors"
	"fmt"
	"jbpf_protobuf_cli/common"
	"jbpf_protobuf_cli/schema"
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type runOptions struct {
	schema  *schema.ClientOptions
	general *common.GeneralOptions

	filePath   string
	inlineJSON string
	payload    string
	streamID   string
	streamUUID uuid.UUID
}

func addToFlags(flags *pflag.FlagSet, opts *runOptions) {
	flags.StringVarP(&opts.filePath, "file", "f", "", "path to file containing payload in JSON format")
	flags.StringVarP(&opts.inlineJSON, "inline-json", "j", "", "inline payload in JSON format")
	flags.StringVar(&opts.streamID, "stream-id", "00000000-0000-0000-0000-000000000000", "stream ID")
}

func (o *runOptions) parse() error {
	if (len(o.inlineJSON) > 0 && len(o.filePath) > 0) || (len(o.inlineJSON) == 0 && len(o.filePath) == 0) {
		return errors.New("exactly one of --file or --inline-json can be specified")
	}

	if len(o.filePath) != 0 {
		if fi, err := os.Stat(o.filePath); err != nil {
			return err
		} else if fi.IsDir() {
			return fmt.Errorf(`expected "%s" to be a file, got a directory`, o.filePath)
		}
		payload, err := os.ReadFile(o.filePath)
		if err != nil {
			return err
		}
		var deserializedPayload interface{}
		err = json.Unmarshal(payload, &deserializedPayload)
		if err != nil {
			return err
		}
		o.payload = string(payload)
	} else {
		var deserializedPayload interface{}
		err := json.Unmarshal([]byte(o.inlineJSON), &deserializedPayload)
		if err != nil {
			return err
		}
		o.payload = o.inlineJSON
	}

	var err error
	o.streamUUID, err = uuid.Parse(o.streamID)
	if err != nil {
		return err
	}

	return nil
}

// Command Load a schema to a local decoder
func Command(opts *common.GeneralOptions) *cobra.Command {
	runOptions := &runOptions{
		schema:  &schema.ClientOptions{},
		general: opts,
	}
	cmd := &cobra.Command{
		Use:   "control",
		Short: "Load a control message via a local decoder",
		Long:  "Load a control message via a local decoder",
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

	return client.SendControl(opts.streamUUID, string(opts.payload))
}
