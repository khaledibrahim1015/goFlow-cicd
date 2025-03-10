package executor

import (
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	maxRetries = 3
	retryDelay = 2 * time.Second
)

func Run(cmd *exec.Cmd) error {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		logrus.Infof("Running (attempt %d/%d): %s", attempt, maxRetries, cmd.String())
		output, err := cmd.CombinedOutput()
		if err == nil {
			logrus.Infof("Output: %s", string(output))
			return nil
		}
		logrus.Errorf("Failed (attempt %d/%d): %s", attempt, maxRetries, string(output))
		if attempt < maxRetries {
			time.Sleep(retryDelay)
		} else {
			return err
		}
	}
	return nil
}
func RunWithOutput(cmd *exec.Cmd) (string, error) {
	output, err := cmd.CombinedOutput()
	return string(output), err
}
