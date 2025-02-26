package common

import (
	"fmt"

	yaml "gopkg.in/yaml.v3"
)

// ProtobufRawConfig represents the configuration for a protobuf message as defined in the yaml config
type ProtobufRawConfig struct {
	MsgName     string `yaml:"msg_name"`
	PackagePath string `yaml:"package_path"`
}

// SerdeRawConfig represents the configuration for serialize/deserialize as defined in the yaml config
type SerdeRawConfig struct {
	Protobuf *ProtobufRawConfig `yaml:"protobuf"`
}

// IOChannelRawConfig represents the configuration for an IO channel as defined in the yaml config
type IOChannelRawConfig struct {
	Serde    *SerdeRawConfig `yaml:"serde"`
	StreamID string          `yaml:"stream_id"`
}

// CodeletDescriptorRawConfig represents the configuration for a codelet descriptor as defined in the yaml config
type CodeletDescriptorRawConfig struct {
	InIOChannel  []*IOChannelRawConfig `yaml:"in_io_channel"`
	OutIOChannel []*IOChannelRawConfig `yaml:"out_io_channel"`
}

// CodeletsetRawConfig represents the configuration for loading a decoder as defined in the yaml config
type CodeletsetRawConfig struct {
	CodeletDescriptor []*CodeletDescriptorRawConfig `yaml:"codelet_descriptor"`
}

func newCodeletsetRawConfig(filePath string) (*CodeletsetRawConfig, error) {
	f, err := NewFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var rawConfig CodeletsetRawConfig
	if err := yaml.Unmarshal(f.Data, &rawConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file %s: %w", filePath, err)
	}

	return &rawConfig, nil
}
