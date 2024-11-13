package unload

import (
	"errors"
	"fmt"
	"jbpf_protobuf_cli/common"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// IOChannelConfig represents the configuration for an IO channel
type IOChannelConfig struct {
	StreamID string `yaml:"stream_id"`

	streamUUID uuid.UUID
}

// CodeletDescriptorConfig represents the configuration for a codelet descriptor
type CodeletDescriptorConfig struct {
	InIOChannel  []*IOChannelConfig `yaml:"in_io_channel"`
	OutIOChannel []*IOChannelConfig `yaml:"out_io_channel"`
}

// DecoderUnloadConfig represents the configuration for unloading a decoder
type DecoderUnloadConfig struct {
	CodeletDescriptor []*CodeletDescriptorConfig `yaml:"codelet_descriptor"`
}

func (io *IOChannelConfig) verify() error {
	streamUUID, err := uuid.Parse(io.StreamID)
	if err != nil {
		return err
	}
	io.streamUUID = streamUUID
	return nil
}

func fromFiles(configs ...string) ([]DecoderUnloadConfig, error) {
	out := make([]DecoderUnloadConfig, 0, len(configs))
	errs := make([]error, 0, len(configs))

configLoad:
	for _, c := range configs {
		f, err := common.NewFile(c)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to read file %s: %w", c, err))
			continue
		}
		var config DecoderUnloadConfig
		if err := yaml.Unmarshal(f.Data, &config); err != nil {
			errs = append(errs, fmt.Errorf("failed to unmarshal file %s: %w", c, err))
			continue
		}

		for _, desc := range config.CodeletDescriptor {
			for _, io := range desc.InIOChannel {
				if err := io.verify(); err != nil {
					errs = append(errs, fmt.Errorf("failed to verify in_io_channel in file %s: %w", c, err))
					continue configLoad
				}
			}
			for _, io := range desc.OutIOChannel {
				if err := io.verify(); err != nil {
					errs = append(errs, fmt.Errorf("failed to verify out_io_channel in file %s: %w", c, err))
					continue configLoad
				}
			}
		}

		out = append(out, config)
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}

	return out, nil
}
