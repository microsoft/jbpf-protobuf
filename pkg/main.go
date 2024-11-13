package main

import (
	"context"
	"jbpf_protobuf_cli/cmd/decoder"
	"jbpf_protobuf_cli/cmd/serde"
	"jbpf_protobuf_cli/common"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	ctx := context.Background()
	if err := cli().ExecuteContext(ctx); err != nil {
		logrus.WithError(err).Fatal("Exiting")
	}
}

func cli() *cobra.Command {
	cmd := &cobra.Command{
		Use:  os.Args[0],
		Long: "jbpf companion command line tool to generate protobuf assets and a local decoder to interact with a remote jbpf instance over sockets.",
	}
	opts := common.NewGeneralOptions(cmd.PersistentFlags())
	cmd.AddCommand(
		decoder.Command(opts),
		serde.Command(opts),
	)
	return cmd
}
