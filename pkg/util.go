package pkg

import (
	"bytes"
	"context"
	"fmt"
	"go.uber.org/zap"
	"os"
	"os/exec"
)

func IsFile(filePath string) error {
	file, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if !file.Mode().IsRegular() {
		return fmt.Errorf("%s is not a file", file)
	}

	return nil
}

// TODO - stdout and stderr as a pipe?
func RunCmd(ctx context.Context, logger *zap.Logger, c string, arg ...string) error {
	cmd := exec.CommandContext(ctx, c, arg...)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	logger.Debug(fmt.Sprintf("running command: %s %v", c, arg))
	err := cmd.Run()
	stdoutStr := stdout.String()
	if stdoutStr != "" {
		logger.Debug(stdoutStr)
	}
	stderrStr := stderr.String()
	if stderrStr != "" {
		logger.Debug(stderrStr)
	}

	if err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			if err.ExitCode() == -1 {
				// Killed by signal
				return nil
			}
		}
		return err
	}

	return nil
}
