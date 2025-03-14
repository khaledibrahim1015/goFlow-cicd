package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/khaledibrahim1015/goFlow-cicd/pkg/executor"
	"github.com/sirupsen/logrus"
)

const (
	dotnet = "dotnet"
	java   = "java"
)

func (p *Pipeline) build() error {
	logrus.Info("Building project with Docker-like behavior...")

	// Define a fixed output directory for .NET (mimics Docker's /app/build or /app/publish)
	buildOutputDir := filepath.Join(p.repoPath, "build-output")
	if p.cfg.Build.Type == dotnet {
		if err := os.MkdirAll(buildOutputDir, 0755); err != nil {
			return fmt.Errorf("failed to create build output directory %s: %v", buildOutputDir, err)
		}
	}

	var buildFile string
	var restoreCmdFunc, buildCmdFunc func() *exec.Cmd
	switch p.cfg.Build.Type {
	case dotnet:
		var err error
		buildFile, err = findBuildFile(p.repoPath, dotnet)
		if err != nil {
			return fmt.Errorf("failed to locate .NET project file: %v", err)
		}
		restoreCmdFunc = func() *exec.Cmd {
			cmd := exec.Command("dotnet", "restore", buildFile)
			cmd.Env = os.Environ()
			cmd.Dir = p.repoPath
			return cmd
		}
		// Mimic Docker's dotnet publish -o
		buildCmdFunc = func() *exec.Cmd {
			cmd := exec.Command("dotnet", "publish", buildFile, "-c", "Release", "-o", buildOutputDir, "/p:UseAppHost=false")
			cmd.Env = os.Environ()
			cmd.Dir = p.repoPath
			return cmd
		}
	case java:
		var err error
		buildFile, err = findBuildFile(p.repoPath, java)
		if err != nil {
			return fmt.Errorf("failed to locate pom.xml: %v", err)
		}
		// Mimic Docker's mvn clean package
		buildCmdFunc = func() *exec.Cmd {
			cmd := exec.Command("mvn", "clean", "package", "-f", buildFile)
			cmd.Env = os.Environ()
			cmd.Dir = p.repoPath
			return cmd
		}
	default:
		logrus.Warnf("Unknown build type '%s', skipping build step", p.cfg.Build.Type)
		return nil
	}

	// Restore for dotnet only
	if p.cfg.Build.Type == dotnet {
		for attempt := 1; attempt <= 3; attempt++ {
			cmd := restoreCmdFunc()
			output, err := executor.RunWithOutput(cmd)
			if err == nil {
				logrus.Infof("Restore output: %s", output)
				break
			}
			logrus.Errorf("Restore failed (attempt %d/3): %v\nOutput: %s", attempt, err, output)
			if attempt == 3 {
				return fmt.Errorf("dotnet restore failed after 3 attempts: %v", err)
			}
		}
	}

	// Build (or publish for dotnet)
	for attempt := 1; attempt <= 3; attempt++ {
		cmd := buildCmdFunc()
		output, err := executor.RunWithOutput(cmd)
		if err == nil {
			logrus.Infof("Build output: %s", output)
			break
		}
		logrus.Errorf("Build failed (attempt %d/3): %v\nOutput: %s", attempt, err, output)
		if attempt == 3 {
			return fmt.Errorf("%s build failed after 3 attempts: %v", p.cfg.Build.Type, err)
		}
	}

	// Ensure output path exists
	if err := os.MkdirAll(p.cfg.Build.OutputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %v", p.cfg.Build.OutputPath, err)
	}

	// Move artifacts
	var srcDir string
	if p.cfg.Build.Type == dotnet {
		srcDir = buildOutputDir
	} else if p.cfg.Build.Type == java {
		srcDir = filepath.Join(filepath.Dir(buildFile), "target")
	}
	if err := moveArtifacts(srcDir, p.cfg.Build.OutputPath); err != nil {
		return fmt.Errorf("failed to move build artifacts: %v", err)
	}

	logrus.Info("Build completed successfully")
	return nil
}

// findBuildFile locates the build file for the given build type
func findBuildFile(dir, buildType string) (string, error) {
	var targetExtensions []string
	switch buildType {
	case dotnet:
		targetExtensions = []string{".csproj"} // Only look for .csproj
	case java:
		targetExtensions = []string{".xml"}
	default:
		return "", fmt.Errorf("unsupported build type: %s", buildType)
	}
	var buildFile string
	err := walkDir(dir, func(path string, info os.DirEntry) error {
		if info.IsDir() {
			return nil
		}
		name := info.Name()
		if buildType == java && name == "pom.xml" {
			buildFile = path
			return filepath.SkipDir
		}
		for _, ext := range targetExtensions {
			if strings.HasSuffix(name, ext) {
				buildFile = path
				return filepath.SkipDir
			}
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("error searching for build file: %v", err)
	}
	if buildFile == "" {
		return "", fmt.Errorf("no build file found for %s in %s", buildType, dir)
	}
	logrus.Infof("Found build file: %s", buildFile)
	return buildFile, nil
}
func walkDir(dir string, fn func(path string, info os.DirEntry) error) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if err := fn(path, entry); err != nil {
			if err == filepath.SkipDir {
				continue
			}
			return err
		}
		if entry.IsDir() {
			if err := walkDir(path, fn); err != nil {
				return err
			}
		}
	}
	return nil
}

// moveArtifacts copies build outputs to the specified output path
func moveArtifacts(srcDir, outputPath string) error {
	// Check if source directory exists
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("build output directory %s not found", srcDir)
	}

	// Log contents of srcDir for debugging
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("failed to read build output directory %s: %v", srcDir, err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("build output directory %s is empty", srcDir)
	}
	logrus.Infof("Found %d files in %s:", len(entries), srcDir)
	for _, entry := range entries {
		logrus.Infof(" - %s", entry.Name())
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %v", outputPath, err)
	}

	// Use absolute paths
	absSrcDir, err := filepath.Abs(srcDir)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path for %s: %v", srcDir, err)
	}
	absOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path for %s: %v", outputPath, err)
	}

	// Run cp through a shell to handle glob expansion
	cmdStr := fmt.Sprintf("cp -r %s/* %s", absSrcDir, absOutputPath)
	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Env = os.Environ()
	output, err := executor.RunWithOutput(cmd)
	if err != nil {
		logrus.Errorf("Command failed: %s", cmdStr)
		return fmt.Errorf("failed to move artifacts from %s to %s: %v\nOutput: %s", absSrcDir, absOutputPath, err, output)
	}
	logrus.Infof("Moved artifacts from %s to %s: %s", absSrcDir, absOutputPath, output)
	return nil
}
