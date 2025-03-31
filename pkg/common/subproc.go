package common

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

// RunSubprocess runs a subprocess
func RunSubprocess(ctx context.Context, logger *logrus.Logger, name string, args ...string) error {
	l := logger.WithFields(logrus.Fields{
		"cmd": strings.Join(append([]string{name}, args...), " "),
	})
	l.Debug("Creating subprocess")
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = os.Environ()
	l.Debug("Running subprocess")
	cmd.Stderr = logger.WithField("channel", "stderr").WriterLevel(logrus.ErrorLevel)
	cmd.Stdout = logger.WithField("channel", "stdout").WriterLevel(logrus.DebugLevel)
	if err := cmd.Run(); err != nil {
		l.Error(fmt.Errorf("failed to run command: %s with error %s", cmd, err))
		return errors.Join(err)
	}
	l.Debug("Complete subprocess")
	return nil
}
