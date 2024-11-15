package main

import (
	"context"
	"jbpf_protobuf_cli/cmd/decoder"
	"jbpf_protobuf_cli/cmd/input"
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
		Long: "jbpf companion command line tool to generate protobuf assets. Includes a decoder to receive output data over a UDP socket from a jbpf instance. Messages are then decoded and print as json. Also provides a mechanism to dispatch input control messages to a jbpf instance via a TCP socket.",
	}
	opts := common.NewGeneralOptions(cmd.PersistentFlags())
	cmd.AddCommand(
		decoder.Command(opts),
		input.Command(opts),
		serde.Command(opts),
	)
	return cmd
}
