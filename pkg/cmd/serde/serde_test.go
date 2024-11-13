package serde

import (
	"errors"
	"fmt"
	"io/fs"
	"jbpf_protobuf_cli/common"
	"log"
	"os"
	"path/filepath"
	"testing"

	_ "embed"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var workdir = os.Getenv("TEST_WORKDIR")
var snapshotdir = os.Getenv("SNAPSHOT_DIR")
var generateSnapshot = os.Getenv("REGENERATE_SNAPSHOT") == "true"
var generalOpts *common.GeneralOptions

func init() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	generalOpts = common.NewGeneralOptionsFromLogger(logger)

	if workdir == "" {
		log.Fatal(`"TEST_WORKDIR" not set`)
	}
	if snapshotdir == "" {
		log.Fatal(`"SNAPSHOT_DIR" not set`)
	}
	err := errors.Join(
		verifyDirExists(workdir, false),
		verifyDirExists(snapshotdir, true),
	)
	if err != nil {
		log.Fatal(err)
	}

}

func verifyDirExists(dir string, createIfNotExists bool) error {
	f, err := os.Stat(dir)
	if err != nil && os.IsNotExist(err) && createIfNotExists {
		if err := os.Mkdir(dir, 0755); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if !f.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}
	return nil
}

func snapshotTest(t *testing.T, snapshotDir, outDir string, cmd *cobra.Command) {
	err := cmd.Execute()
	require.NoError(t, err)

	cFiles, err := filepath.Glob(outDir + "/*.c")
	require.NoError(t, err)
	hFiles, err := filepath.Glob(outDir + "/*.h")
	require.NoError(t, err)
	pbFiles, err := filepath.Glob(outDir + "/*.pb")
	require.NoError(t, err)

	outDirFiles := append(cFiles, hFiles...)
	outDirFiles = append(outDirFiles, pbFiles...)

	for _, file := range outDirFiles {
		baseName := filepath.Base(file)
		snapshotFile := filepath.Join(snapshotDir, baseName)

		if generateSnapshot {
			require.NoError(t, moveFile(file, snapshotFile))
		} else {
			newFile, err := os.ReadFile(file)
			require.NoError(t, err)
			snapshotFile, err := os.ReadFile(snapshotFile)
			require.NoError(t, err)
			assert.Equal(t, snapshotFile, newFile, "file %s does not match snapshot", baseName)
		}
	}
}

func moveFile(source, destination string) error {
	fi, err := os.Stat(source)
	if err != nil {
		return err
	} else if fi.IsDir() {
		return fmt.Errorf("expected %s to be a file, got dir", source)
	}
	fileMod := fi.Mode()
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if err := os.Remove(source); err != nil {
		return err
	}

	newFi, err := os.Stat(destination)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	} else if err == nil && newFi.IsDir() {
		return fmt.Errorf("expected %s to be a file, got dir", destination)
	} else if err == nil {
		if err := os.Remove(destination); err != nil {
			return err
		}
	}
	destFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer func() {
		if err := destFile.Close(); err != nil {
			fmt.Printf("failed to close destination file %s when moving\n", destination)
		}
	}()

	n, err := destFile.Write(data)
	if n != len(data) {
		return fmt.Errorf("failed to write entire file %s, wrote %d of %d bytes, with err %v", destination, n, len(data), err)
	} else if err != nil {
		return err
	}

	return os.Chmod(destination, fileMod)
}

func TestCases(t *testing.T) {
	testArgs := map[string][]string{
		"example1": {"-s", "example:req_resp,status", "-w", filepath.Join(workdir, "example1")},
		"example2": {"-s", "example2:item", "-w", filepath.Join(workdir, "example2")},
		"example3": {"-s", "example3:obj", "-w", filepath.Join(workdir, "example3")},
	}

	for exampleName, testArgs := range testArgs {
		t.Run(exampleName, func(t *testing.T) {
			outDir, err := os.MkdirTemp("", exampleName)
			require.NoError(t, err)
			defer func() {
				if err := os.RemoveAll(outDir); err != nil {
					t.Logf("failed to remove outDir: %s", err)
				}
			}()
			snapshotDir := filepath.Join(snapshotdir, exampleName)
			err = verifyDirExists(snapshotDir, true)
			require.NoError(t, err)
			cmd := Command(generalOpts)
			cmd.SetArgs(append(testArgs, "-o", outDir))
			snapshotTest(t, snapshotDir, outDir, cmd)
		})
	}
}
