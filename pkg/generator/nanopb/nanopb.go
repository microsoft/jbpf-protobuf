package nanopb

import (
	"fmt"
	"log"
	"os"
)

const (
	nanoPbEnvVar = "NANO_PB"
)

var (
	// GeneratorPath is $NANO_PB/generator/nanopb_generator
	GeneratorPath string
	// ProtocPath is $NANO_PB/generator/protoc
	ProtocPath string
	// Path is $NANO_PB
	Path string
	// PbCommonCPath is $NANO_PB/pb_common.c
	PbCommonCPath string
	// PbDecodeCPath is $NANO_PB/pb_decode.c
	PbDecodeCPath string
	// PbEncodeCPath is $NANO_PB/pb_encode.c
	PbEncodeCPath string
)

func init() {
	Path = os.Getenv(nanoPbEnvVar)

	if err := validateDirPath(Path); err != nil {
		log.Fatal(err)
	}

	ProtocPath = fmt.Sprintf("%s/generator/protoc", Path)
	GeneratorPath = fmt.Sprintf("%s/generator/nanopb_generator", Path)
	PbCommonCPath = fmt.Sprintf("%s/pb_common.c", Path)
	PbDecodeCPath = fmt.Sprintf("%s/pb_decode.c", Path)
	PbEncodeCPath = fmt.Sprintf("%s/pb_encode.c", Path)
}

func validateDirPath(path string) error {
	if path == "" {
		return nil
	}
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf(`Expected "%s" to be a directory`, path)
	}
	return nil
}
