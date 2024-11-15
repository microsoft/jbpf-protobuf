package schema

import (
	"context"
	"errors"
	"fmt"
	"jbpf_protobuf_cli/common"
	"jbpf_protobuf_cli/generator/nanopb"
	"jbpf_protobuf_cli/generator/stream"
	"os"

	"github.com/sirupsen/logrus"
)

const (
	pbTemplate = "%s.pb"
)

var (
	generatedFileTemplate = []string{pbTemplate, "%s.pb.c", "%s.pb.h"}
)

// Config for schema file generation
type Config struct {
	Files             []*common.File
	ProtoMessageNames []string
	ProtoPackageName  string
}

// Generate generates files for schema inside a temporary directory
func Generate(ctx context.Context, logger *logrus.Logger, cfg *Config) ([]*common.File, error) {
	wd, err := os.MkdirTemp("", "temp*")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = os.RemoveAll(wd); err != nil {
			logger.WithField("directory", wd).WithError(err).Error("failed to remove working directory")
		}
	}()

	originalWd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if err := os.Chdir(wd); err != nil {
		return nil, err
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			logger.WithField("directory", wd).WithError(err).Error("failed to change working directory")
		}
	}()

	for _, fileDetails := range cfg.Files {
		logger.Debug("Writing file: ", fileDetails.Name)
		f, err := os.Create(fileDetails.Name)
		if err != nil {
			return nil, err
		}
		n, err := f.Write(fileDetails.Data)
		if n != len(fileDetails.Data) {
			err = errors.Join(err, fmt.Errorf("expected to write %d bytes, wrote %d", len(fileDetails.Data), n))
		}
		if err != nil {
			return nil, errors.Join(err, f.Close())
		}
		logger.Debug("Closed file: ", fileDetails.Name)
		if err = f.Close(); err != nil {
			return nil, err
		}
	}

	if err := errors.Join(
		common.RunSubprocess(
			ctx,
			logger,
			nanopb.GeneratorPath,
			cfg.ProtoPackageName+".proto",
		),
		common.RunSubprocess(
			ctx,
			logger,
			nanopb.ProtocPath,
			cfg.ProtoPackageName+".proto",
			"-o",
			fmt.Sprintf(pbTemplate, cfg.ProtoPackageName),
		)); err != nil {
		return nil, err
	}

	generatedFiles := make([]*common.File, 0, len(cfg.ProtoMessageNames)*2+3)

	for _, fTemplate := range generatedFileTemplate {
		f := fmt.Sprintf(fTemplate, cfg.ProtoPackageName)
		fileData, err := common.NewFile(f)
		if err != nil {
			return nil, err
		}
		generatedFiles = append(generatedFiles, fileData)
	}

	for _, protoMessageName := range cfg.ProtoMessageNames {
		files, err := stream.Generate(ctx, logger, cfg.ProtoPackageName, protoMessageName)
		if err != nil {
			return nil, err
		}
		generatedFiles = append(generatedFiles, files...)
	}

	return generatedFiles, nil
}
