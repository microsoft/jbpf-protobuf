package common

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// File represents a generated file
type File struct {
	Data []byte
	Mode fs.FileMode
	Name string
}

// NewFile creates a new file from a file path
func NewFile(filePath string) (*File, error) {
	filePath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	} else if fi.IsDir() {
		return nil, fmt.Errorf(`expected "%s" to be a file, got a directory`, filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return &File{
		Data: content,
		Mode: fi.Mode(),
		Name: filepath.Base(filePath),
	}, nil
}

// WriteFilesToDirectory writes files to a directory
func WriteFilesToDirectory(logger *logrus.Logger, outputDirectory string, files []*File) error {
	for _, f := range files {
		if err := WriteFileToDirectory(logger, outputDirectory, f); err != nil {
			return err
		}
	}
	return nil
}

// WriteFileToDirectory writes a file to a directory
func WriteFileToDirectory(logger *logrus.Logger, outputDirectory string, file *File) error {
	filePath := filepath.Join(outputDirectory, file.Name)

	l := logger.WithField("filename", filePath)
	fi, err := os.Stat(filePath)
	var f *os.File
	if err == nil && fi.IsDir() {
		return fmt.Errorf(`"%s" is a directory`, filePath)
	} else if !os.IsNotExist(err) {
		l.Debug("Overwriting existing file")
		if err := os.Remove(filePath); err != nil {
			return err
		}
	} else if err == nil {
		l.Debug("Creating file")
	}
	f, err = os.Create(filePath)
	if err != nil {
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			l.WithError(err).Error("failed to close file")
		}
	}()

	n, err := f.Write(file.Data)
	if err != nil {
		return err
	} else if n != len(file.Data) {
		return fmt.Errorf("expected to write %d bytes, wrote %d", len(file.Data), n)
	}

	if err := os.Chmod(filePath, file.Mode); err != nil {
		return err
	}

	return nil
}
