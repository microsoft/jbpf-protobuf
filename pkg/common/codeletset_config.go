package common

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// ProtobufConfig represents the configuration for a protobuf message
type ProtobufConfig struct {
	MsgName     string
	PackageName string
	PackagePath string
}

func newProtobufConfig(cfg *ProtobufRawConfig) (*ProtobufConfig, error) {
	if len(cfg.MsgName) == 0 {
		return nil, fmt.Errorf("missing required field serde.protobuf.msg_name")
	}
	if len(cfg.PackagePath) == 0 {
		return nil, fmt.Errorf("missing required field serde.protobuf.package_path")
	}

	packagePath := os.ExpandEnv(cfg.PackagePath)
	basename := filepath.Base(packagePath)

	return &ProtobufConfig{
		MsgName:     cfg.MsgName,
		PackageName: strings.TrimSuffix(basename, filepath.Ext(basename)),
		PackagePath: packagePath,
	}, nil
}

// SerdeConfig represents the configuration for serialize/deserialize
type SerdeConfig struct {
	Protobuf *ProtobufConfig
}

func newSerdeConfig(cfg *SerdeRawConfig) (*SerdeConfig, error) {
	if cfg.Protobuf == nil {
		return nil, fmt.Errorf("missing required field serde.protobuf")
	}

	protobuf, err := newProtobufConfig(cfg.Protobuf)
	if err != nil {
		return nil, err
	}

	return &SerdeConfig{Protobuf: protobuf}, nil
}

// IOChannelConfig represents the configuration for an IO channel
type IOChannelConfig struct {
	Serde      *SerdeConfig
	StreamUUID uuid.UUID
}

func newIOChannelConfig(cfg *IOChannelRawConfig) (*IOChannelConfig, error) {
	if cfg.Serde == nil {
		return nil, fmt.Errorf("missing required field serde")
	}

	serde, err := newSerdeConfig(cfg.Serde)
	if err != nil {
		return nil, err
	}

	streamUUID, err := uuid.Parse(cfg.StreamID)
	if err != nil {
		return nil, err
	}

	return &IOChannelConfig{
		Serde:      serde,
		StreamUUID: streamUUID,
	}, nil
}

// CodeletDescriptorConfig represents the configuration for a codelet descriptor
type CodeletDescriptorConfig struct {
	InIOChannel  []*IOChannelConfig
	OutIOChannel []*IOChannelConfig
}

func newCodeletDescriptorConfig(cfg *CodeletDescriptorRawConfig) (*CodeletDescriptorConfig, error) {
	inIOChannel := make([]*IOChannelConfig, 0, len(cfg.InIOChannel))
	for _, rawIO := range cfg.InIOChannel {
		io, err := newIOChannelConfig(rawIO)
		if err != nil {
			return nil, err
		}
		inIOChannel = append(inIOChannel, io)
	}

	outIOChannel := make([]*IOChannelConfig, 0, len(cfg.OutIOChannel))
	for _, rawIO := range cfg.OutIOChannel {
		io, err := newIOChannelConfig(rawIO)
		if err != nil {
			return nil, err
		}
		outIOChannel = append(outIOChannel, io)
	}

	return &CodeletDescriptorConfig{
		InIOChannel:  inIOChannel,
		OutIOChannel: outIOChannel,
	}, nil
}

// CodeletsetConfig represents the configuration for loading a decoder
type CodeletsetConfig struct {
	CodeletDescriptor []*CodeletDescriptorConfig
}

func newCodeletSetConfig(cfg *CodeletsetRawConfig) (*CodeletsetConfig, error) {
	codeletDescriptors := make([]*CodeletDescriptorConfig, 0, len(cfg.CodeletDescriptor))
	for _, rawDesc := range cfg.CodeletDescriptor {
		desc, err := newCodeletDescriptorConfig(rawDesc)
		if err != nil {
			return nil, err
		}
		codeletDescriptors = append(codeletDescriptors, desc)
	}
	return &CodeletsetConfig{CodeletDescriptor: codeletDescriptors}, nil
}

// LoadCompiledProtos loads the compiled protobuf files from the codeletset config
func LoadCompiledProtos(cfgs []*CodeletsetConfig, includeInIO, includeOutIO bool) (map[string]*File, error) {
	compiledProtos := make(map[string]*File)

	for _, c := range cfgs {
		for _, desc := range c.CodeletDescriptor {
			if includeInIO {
				for _, io := range desc.InIOChannel {
					if _, ok := compiledProtos[io.Serde.Protobuf.PackagePath]; !ok {
						protoPkg, err := NewFile(io.Serde.Protobuf.PackagePath)
						if err != nil {
							return nil, err
						}
						compiledProtos[io.Serde.Protobuf.PackagePath] = protoPkg
					}
				}
			}

			if includeOutIO {
				for _, io := range desc.OutIOChannel {
					if _, ok := compiledProtos[io.Serde.Protobuf.PackagePath]; !ok {
						protoPkg, err := NewFile(io.Serde.Protobuf.PackagePath)
						if err != nil {
							return nil, err
						}
						compiledProtos[io.Serde.Protobuf.PackagePath] = protoPkg
					}
				}
			}
		}
	}

	return compiledProtos, nil
}

// CodeletsetConfigFromFiles reads and unmarshals the given files into a slice of CodeletsetConfig
func CodeletsetConfigFromFiles(configs ...string) ([]*CodeletsetConfig, error) {
	out := make([]*CodeletsetConfig, 0, len(configs))
	errs := make([]error, 0, len(configs))

	for _, c := range configs {
		rawConfig, err := newCodeletsetRawConfig(c)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		config, err := newCodeletSetConfig(rawConfig)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to unpack file %s: %w", c, err))
			continue
		}

		out = append(out, config)
	}

	return out, errors.Join(errs...)
}
