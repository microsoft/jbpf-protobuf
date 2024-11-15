package nanopb

import (
	"errors"
	"jbpf_protobuf_cli/common"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	optionsGlob = "*.options"
	protosGlob  = "*.proto"
)

// FindFiles finds nanopb files (*.proto, *.options) in a directory
func FindFiles(logger *logrus.Logger, workingDir string) ([]*common.File, error) {
	optionsFiles, err1 := filepath.Glob(filepath.Join(workingDir, optionsGlob))
	protoFiles, err2 := filepath.Glob(filepath.Join(workingDir, protosGlob))
	if err := errors.Join(err1, err2); err != nil {
		return nil, err
	}

	files := make([]*common.File, 0, len(optionsFiles)+len(protoFiles))
	fileNames := make([]string, 0, len(optionsFiles)+len(protoFiles))

	for _, f := range optionsFiles {
		file, err := common.NewFile(f)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
		fileNames = append(fileNames, filepath.Base(f))
	}
	for _, f := range protoFiles {
		file, err := common.NewFile(f)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
		fileNames = append(fileNames, filepath.Base(f))
	}

	if len(files) == 0 {
		return nil, errors.New("no nanopb files found")
	} else if len(protoFiles) == 0 {
		return nil, errors.New("no proto file found")
	}

	logger.WithField("files", strings.Join(fileNames, ", ")).Debug("found nanopb files")

	return files, nil
}
