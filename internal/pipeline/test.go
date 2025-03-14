package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/khaledibrahim1015/goFlow-cicd/pkg/executor"
	"github.com/sirupsen/logrus"
)

func (p *Pipeline) test() error {
	logrus.Info("Starting test stage...")

	if p.cfg.Test.Type == "" {
		logrus.Warn("No test configuration specified, skipping test stage")
		return nil
	}

	testOutputDir := filepath.Join(p.repoPath, "test-output")
	if err := os.MkdirAll(testOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create test output directory %s: %v", testOutputDir, err)
	}

	var testFile string
	var testCmdFunc func() *exec.Cmd
	switch p.cfg.Test.Type {
	case dotnet:
		var err error
		testFile, err = findBuildFile(p.repoPath, dotnet)
		if err != nil {
			return fmt.Errorf("failed to locate .NET test project file: %v", err)
		}
		logrus.Infof("Running .NET tests for %s", testFile)
		testCmdFunc = func() *exec.Cmd {
			cmd := exec.Command("dotnet", "test", testFile, "--configuration", "Release",
				"--logger", "trx", "--results-directory", testOutputDir)
			cmd.Env = append(os.Environ(), "DOTNET_CLI_TELEMETRY_OPTOUT=1")
			cmd.Dir = p.repoPath
			return cmd
		}
	case java:
		var err error
		testFile, err = findBuildFile(p.repoPath, java)
		if err != nil {
			return fmt.Errorf("failed to locate pom.xml for tests: %v", err)
		}
		logrus.Infof("Running Java tests with Maven: %s", testFile)
		testCmdFunc = func() *exec.Cmd {
			cmd := exec.Command("mvn", "test", "-f", testFile)
			cmd.Env = os.Environ()
			cmd.Dir = p.repoPath
			return cmd
		}
	default:
		logrus.Warnf("Unknown test type '%s', skipping test step", p.cfg.Test.Type)
		return nil
	}

	for attempt := 1; attempt <= 3; attempt++ {
		cmd := testCmdFunc()
		output, err := executor.RunWithOutput(cmd)
		if err == nil {
			logrus.Debugf("Test succeeded: %s", output)
			break
		}
		logrus.Errorf("Test failed (attempt %d/3): %v\nOutput: %s", attempt, err, output)
		if attempt == 3 {
			return fmt.Errorf("%s test failed after 3 attempts: %v", p.cfg.Test.Type, err)
		}
	}

	// Move test reports
	reportOutputPath := p.cfg.Test.OutputPath
	if reportOutputPath == "" {
		reportOutputPath = filepath.Join(p.cfg.Build.OutputPath, "test-reports")
	}
	if err := os.MkdirAll(reportOutputPath, 0755); err != nil {
		return fmt.Errorf("failed to create test report directory %s: %v", reportOutputPath, err)
	}
	if err := moveArtifacts(testOutputDir, reportOutputPath); err != nil {
		return fmt.Errorf("failed to move test reports: %v", err)
	}

	logrus.Info("Test stage completed successfully")
	return nil
}
