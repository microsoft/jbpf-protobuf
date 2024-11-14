package common

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// ProtobufConfig represents the configuration for a protobuf message
type ProtobufConfig struct {
	AbsPackagePath   string `yaml:"-"`
	MsgName          string `yaml:"msg_name"`
	PackagePath      string `yaml:"package_path"`
	ProtoPackageName string `yaml:"-"`
}

// SerdeConfig represents the configuration for serialize/deserialize
type SerdeConfig struct {
	Protobuf *ProtobufConfig `yaml:"protobuf"`
}

// IOChannelConfig represents the configuration for an IO channel
type IOChannelConfig struct {
	Serde      *SerdeConfig `yaml:"serde"`
	StreamID   string       `yaml:"stream_id"`
	StreamUUID uuid.UUID    `yaml:"-"`
}

// CodeletDescriptorConfig represents the configuration for a codelet descriptor
type CodeletDescriptorConfig struct {
	InIOChannel  []*IOChannelConfig `yaml:"in_io_channel"`
	OutIOChannel []*IOChannelConfig `yaml:"out_io_channel"`
}

// CodeletsetConfig represents the configuration for loading a decoder
type CodeletsetConfig struct {
	CodeletDescriptor []*CodeletDescriptorConfig `yaml:"codelet_descriptor"`
}

func (io *IOChannelConfig) verify(compiledProtos map[string]*File) error {
	streamUUID, err := uuid.Parse(io.StreamID)
	if err != nil {
		return err
	}
	io.StreamUUID = streamUUID
	if io.Serde == nil || io.Serde.Protobuf == nil || io.Serde.Protobuf.PackagePath == "" {
		return fmt.Errorf("missing required field package_path")
	}

	io.Serde.Protobuf.AbsPackagePath = os.ExpandEnv(io.Serde.Protobuf.PackagePath)
	basename := filepath.Base(io.Serde.Protobuf.AbsPackagePath)
	io.Serde.Protobuf.ProtoPackageName = strings.TrimSuffix(basename, filepath.Ext(basename))

	if _, ok := compiledProtos[io.Serde.Protobuf.AbsPackagePath]; !ok {
		protoPkg, err := NewFile(io.Serde.Protobuf.AbsPackagePath)
		if err != nil {
			return err
		}
		compiledProtos[io.Serde.Protobuf.AbsPackagePath] = protoPkg
	}

	return nil
}

// CodeletsetConfigFromFiles reads and unmarshals the given files into a slice of CodeletsetConfig
func CodeletsetConfigFromFiles(configs ...string) ([]CodeletsetConfig, map[string]*File, error) {
	out := make([]CodeletsetConfig, 0, len(configs))
	compiledProtos := make(map[string]*File)
	errs := make([]error, 0, len(configs))

configLoad:
	for _, c := range configs {
		f, err := NewFile(c)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to read file %s: %w", c, err))
			continue
		}
		var config CodeletsetConfig
		if err := yaml.Unmarshal(f.Data, &config); err != nil {
			errs = append(errs, fmt.Errorf("failed to unmarshal file %s: %w", c, err))
			continue
		}

		for _, desc := range config.CodeletDescriptor {
			for _, io := range desc.InIOChannel {
				if err := io.verify(compiledProtos); err != nil {
					errs = append(errs, fmt.Errorf("failed to verify in_io_channel in file %s: %w", c, err))
					continue configLoad
				}
			}
			for _, io := range desc.OutIOChannel {
				if err := io.verify(compiledProtos); err != nil {
					errs = append(errs, fmt.Errorf("failed to verify out_io_channel in file %s: %w", c, err))
					continue configLoad
				}
			}
		}

		out = append(out, config)
	}
	if err := errors.Join(errs...); err != nil {
		return nil, nil, err
	}

	return out, compiledProtos, nil
}
