package serde

import (
	"errors"
	"fmt"
	"jbpf_protobuf_cli/common"
	"jbpf_protobuf_cli/generator/nanopb"
	"jbpf_protobuf_cli/generator/schema"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	relativeWorkingDir = "./"
)

var (
	originalBaseDir string
)

type parsedProtoConfig struct {
	protoPackageName  string
	protoMessageNames []string
}

type runOptions struct {
	general *common.GeneralOptions

	absOutputDir  string
	absWorkingDir string
	outputDir     string
	protoConfigs  []string
	schemas       []*parsedProtoConfig
	workingDir    string
}

func init() {
	var err error
	originalBaseDir, err = filepath.Abs(originalBaseDir)
	if err != nil {
		log.Fatal(err)
	}
}

func addToFlags(flags *pflag.FlagSet, opts *runOptions) {
	flags.StringArrayVarP(&opts.protoConfigs, "schema", "s", []string{}, `source proto file(s), along with any message names. In the form "{proto package name}:{proto message names,}"`)
	flags.StringVarP(&opts.outputDir, "output-dir", "o", relativeWorkingDir, "output directory, will default to the current directory")
	flags.StringVarP(&opts.workingDir, "workdir", "w", relativeWorkingDir, "working directory, will default to the current directory")
}

func validateDir(absPath string) error {
	fi, err := os.Stat(absPath)
	if err != nil {
		return err
	} else if !fi.IsDir() {
		return fmt.Errorf(`Expected "%s" to be a directory`, absPath)
	}
	return nil
}

func (o *runOptions) parse() error {
	var err1, err2 error
	o.absOutputDir, err1 = filepath.Abs(o.outputDir)
	o.absWorkingDir, err2 = filepath.Abs(o.workingDir)

	if err := errors.Join(err1, err2); err != nil {
		return err
	}

	if err := errors.Join(validateDir(o.absOutputDir), validateDir(o.absWorkingDir)); err != nil {
		return err
	}

	o.schemas = make([]*parsedProtoConfig, len(o.protoConfigs))
	for i, s := range o.protoConfigs {
		parts := strings.Split(s, ":")
		if len(parts) != 2 {
			return errors.New("invalid schema format")
		}
		protoPackageName := strings.TrimSpace(parts[0])
		if len(protoPackageName) == 0 {
			return errors.New("invalid schema format")
		}
		protoMessageNames := make([]string, 0)
		if len(parts[1]) > 0 {
			protoMessageNames = strings.Split(parts[1], ",")
			for i := range protoMessageNames {
				protoMessageNames[i] = strings.TrimSpace(protoMessageNames[i])
				if len(protoMessageNames[i]) == 0 {
					return errors.New("invalid schema format")
				}
			}
		}

		o.schemas[i] = &parsedProtoConfig{
			protoPackageName:  protoPackageName,
			protoMessageNames: protoMessageNames,
		}
	}

	return nil
}

// Command Generate serde assets for protobuf spec
func Command(opts *common.GeneralOptions) *cobra.Command {
	runOptions := &runOptions{
		general: opts,
	}
	cmd := &cobra.Command{
		Use:   "serde",
		Short: "Generate serde assets for protobuf spec",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd, runOptions)
		},
		SilenceUsage: true,
	}
	addToFlags(cmd.PersistentFlags(), runOptions)
	return cmd
}

func run(cmd *cobra.Command, opts *runOptions) error {
	if err := errors.Join(
		opts.general.Parse(),
		opts.parse(),
	); err != nil {
		return err
	}

	logger := opts.general.Logger

	for _, cfg := range opts.schemas {
		fileCfgs, err := nanopb.FindFiles(logger, opts.absWorkingDir)
		if err != nil {
			return err
		}

		files, err := schema.Generate(cmd.Context(), logger, &schema.Config{
			Files:             fileCfgs,
			ProtoPackageName:  cfg.protoPackageName,
			ProtoMessageNames: cfg.protoMessageNames,
		})
		if err != nil {
			return err
		}

		if err := common.WriteFilesToDirectory(logger, opts.absOutputDir, files); err != nil {
			return err
		}
	}

	return nil
}
