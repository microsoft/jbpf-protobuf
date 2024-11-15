// Copyright (c) Microsoft Corporation. All rights reserved.

package input

import (
	"jbpf_protobuf_cli/cmd/input/forward"
	"jbpf_protobuf_cli/common"

	"github.com/spf13/cobra"
)

// Command returns the input commands
func Command(opts *common.GeneralOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "input",
		Long:  "Execute a jbpf input subcommand.",
		Short: "Execute a jbpf input subcommand",
	}
	cmd.AddCommand(
		forward.Command(opts),
	)
	return cmd
}
