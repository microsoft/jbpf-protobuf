// Copyright (c) Microsoft Corporation. All rights reserved.

package decoder

import (
	"jbpf_protobuf_cli/cmd/decoder/control"
	"jbpf_protobuf_cli/cmd/decoder/load"
	"jbpf_protobuf_cli/cmd/decoder/run"
	"jbpf_protobuf_cli/cmd/decoder/unload"
	"jbpf_protobuf_cli/common"

	"github.com/spf13/cobra"
)

// Command returns the decoder commands
func Command(opts *common.GeneralOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decoder",
		Long:  "Execute a decoder subcommand.",
		Short: "Execute a decoder subcommand",
	}
	cmd.AddCommand(
		control.Command(opts),
		load.Command(opts),
		unload.Command(opts),
		run.Command(opts),
	)
	return cmd
}
