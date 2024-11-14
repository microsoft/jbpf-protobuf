// Copyright (c) Microsoft Corporation. All rights reserved.

package forward

import (
	"encoding/json"
	"errors"
	"fmt"
	"jbpf_protobuf_cli/common"
	"jbpf_protobuf_cli/jbpf"
	"os"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

type runOptions struct {
	jbpf    *jbpf.Options
	general *common.GeneralOptions

	compiledProtos map[string]*common.File
	configFiles    []string
	configs        []common.CodeletsetConfig
	filePath       string
	inlineJSON     string
	payload        string
	streamID       string
	streamUUID     uuid.UUID
}

func addToFlags(flags *pflag.FlagSet, opts *runOptions) {
	flags.StringArrayVarP(&opts.configFiles, "config", "c", []string{}, "configuration files to load")
	flags.StringVar(&opts.streamID, "stream-id", "00000000-0000-0000-0000-000000000000", "stream ID")
	flags.StringVarP(&opts.filePath, "file", "f", "", "path to file containing payload in JSON format")
	flags.StringVarP(&opts.inlineJSON, "inline-json", "j", "", "inline payload in JSON format")
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

	configs, compiledProtos, err := common.CodeletsetConfigFromFiles(o.configFiles...)
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
		jbpf:    &jbpf.Options{},
		general: opts,
	}
	cmd := &cobra.Command{
		Use:   "forward",
		Short: "Load a control message",
		Long:  "Load a control message",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd, runOptions)
		},
		SilenceUsage: true,
	}
	addToFlags(cmd.PersistentFlags(), runOptions)
	jbpf.AddOptionsToFlags(cmd.PersistentFlags(), runOptions.jbpf)
	return cmd
}

func run(_ *cobra.Command, opts *runOptions) error {
	if err := errors.Join(
		opts.general.Parse(),
		opts.jbpf.Parse(),
		opts.parse(),
	); err != nil {
		return err
	}

	logger := opts.general.Logger

	client, err := jbpf.NewClient(logger, opts.jbpf)
	if err != nil {
		return err
	}

	msg, err := getMessageInstance(opts.configs, opts.compiledProtos, opts.streamUUID)
	if err != nil {
		return err
	}

	err = protojson.Unmarshal([]byte(opts.payload), msg)
	if err != nil {
		logger.WithError(err).Error("error unmarshalling payload")
		return err
	}

	logger.WithFields(logrus.Fields{
		"msg": fmt.Sprintf("%T - \"%v\"", msg, msg),
	}).Info("sending msg")

	payload, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	out := append(opts.streamUUID[:], payload...)

	return client.Write(out)
}

func getMessageInstance(configs []common.CodeletsetConfig, compiledProtos map[string]*common.File, streamUUID uuid.UUID) (*dynamicpb.Message, error) {
	for _, config := range configs {
		for _, desc := range config.CodeletDescriptor {
			for _, io := range desc.InIOChannel {
				fmt.Printf("%v == %v = %v\n", io.StreamUUID, streamUUID, io.StreamUUID == streamUUID)
				if io.StreamUUID == streamUUID {
					compiledProto := compiledProtos[io.Serde.Protobuf.AbsPackagePath]

					fds := &descriptorpb.FileDescriptorSet{}
					if err := proto.Unmarshal(compiledProto.Data, fds); err != nil {
						return nil, err
					}

					pd, err := protodesc.NewFiles(fds)
					if err != nil {
						return nil, err
					}

					msgName := protoreflect.FullName(io.Serde.Protobuf.MsgName)
					var desc protoreflect.Descriptor
					desc, err = pd.FindDescriptorByName(msgName)
					if err != nil {
						return nil, err
					}

					md, ok := desc.(protoreflect.MessageDescriptor)
					if !ok {
						return nil, fmt.Errorf("failed to cast desc to protoreflect.MessageDescriptor, got %T", desc)
					}

					return dynamicpb.NewMessage(md), nil
				}
			}
		}
	}

	return nil, fmt.Errorf("stream %s not found in any of the loaded schemas", streamUUID)
}
