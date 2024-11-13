package stream

import (
	"context"
	"errors"
	"fmt"
	"jbpf_protobuf_cli/common"
	"jbpf_protobuf_cli/generator/nanopb"
	"os"
	"text/template"

	"github.com/sirupsen/logrus"
)

const (
	defaultPbField32Bit       = "1"
	envVarPbField32Bit        = "PB_FIELD_32BIT"
	envVarPbMaxRequiredFields = "PB_MAX_REQUIRED_FIELDS"
	serializerC               = "%s:%s_serializer.c"
	serializerSO              = "%s:%s_serializer.so"
)

func createNewFileWithTmpl(logger *logrus.Logger, filename string, tmpl *template.Template, data SerializerTemplateData) error {
	l := logger.WithField("filename", filename)

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			l.WithError(err).Error("Failed to close file")
		}
	}()
	l.Debug("Created file")
	err = tmpl.Execute(f, data)
	if err != nil {
		l.WithError(err).Error("Failed to write to file")
		return err
	}
	l.Debug("Successfully written to file")
	return nil
}

// Generate creates files for a stream
func Generate(ctx context.Context, logger *logrus.Logger, protoPackageName, protoMessageName string) ([]*common.File, error) {
	cFile := fmt.Sprintf(serializerC, protoPackageName, protoMessageName)
	soFile := fmt.Sprintf(serializerSO, protoPackageName, protoMessageName)

	if err := createNewFileWithTmpl(logger,
		cFile,
		serializerTemplate,
		SerializerTemplateData{ProtoPackageName: protoPackageName, ProtoMessageName: protoMessageName},
	); err != nil {
		return nil, err
	}

	pbField32Bit := os.Getenv(envVarPbField32Bit)
	if pbField32Bit == "" {
		pbField32Bit = defaultPbField32Bit
	}
	args := []string{
		"-I",
		nanopb.Path,
		cFile,
		protoPackageName + ".pb.c",
		nanopb.PbCommonCPath,
		nanopb.PbDecodeCPath,
		nanopb.PbEncodeCPath,
		"-DPB_FIELD_32BIT=" + pbField32Bit,
	}
	if pbMaxRequiredFields := os.Getenv(envVarPbMaxRequiredFields); len(pbMaxRequiredFields) > 0 {
		args = append(args, "-DPB_MAX_REQUIRED_FIELDS="+pbMaxRequiredFields)
	}
	args = append(args, "-shared", "-fPIC", "-o", soFile)

	if err := common.RunSubprocess(ctx, logger, "cc", args...); err != nil {
		return nil, err
	}

	cFileData, err1 := common.NewFile(cFile)
	soFileData, err2 := common.NewFile(soFile)
	if err := errors.Join(err1, err2); err != nil {
		return nil, err
	}

	return []*common.File{cFileData, soFileData}, nil
}
