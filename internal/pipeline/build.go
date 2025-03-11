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

	logrus.Info("building ..................")

	// Stage 1: Build the project
	var buildFile string

	switch p.cfg.Build.Type {
	case dotnet:
		var err error
		buildFile, err = findBuildFile(p.repoPath, "dotnet")
		if err != nil {
			return fmt.Errorf("failed to locate .NET project file: %v", err)
		}
		cmd := exec.Command("dotnet", "build", buildFile)
		if err := executor.Run(cmd); err != nil {
			return fmt.Errorf("dotnet build failed: %v", err)
		}
	case java:
		var err error
		buildFile, err = findBuildFile(p.repoPath, "java")
		if err != nil {
			return fmt.Errorf("failed to locate pom.xml: %v", err)
		}
		cmd := exec.Command("mvn", "clean", "install", "-f", buildFile)
		if err := executor.Run(cmd); err != nil {
			return fmt.Errorf("mvn build failed: %v", err)
		}
	default:
		logrus.Warnf("Unknown build type '%s', skipping build step", p.cfg.Build.Type)
		return nil
	}
	// Stage 2: Move artifacts to output_path
	if err := moveArtifacts(p.cfg.Build.Type, p.repoPath, p.cfg.Build.OutputPath); err != nil {
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
		targetExtensions = []string{".csproj", ".sln"}
	case java:
		targetExtensions = []string{".xml"} // For pom.xml
	default:
		return "", fmt.Errorf("unsupported build type: %s", buildType)
	}
	var buildFile string
	err := walkDir(dir, func(path string, info os.DirEntry) error {
		if info.IsDir() {
			return nil
		}
		name := info.Name()
		if buildType == "java" && name == "pom.xml" {
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
func moveArtifacts(buildType, repoPath, outputPath string) error {
	var srcDir string
	switch buildType {
	case "dotnet":
		// .NET typically outputs to bin/ in the project directory
		buildFile, err := findBuildFile(repoPath, "dotnet")
		if err != nil {
			return err
		}
		srcDir = filepath.Join(filepath.Dir(buildFile), "bin")
	case "java":
		// Maven outputs to target/ in the directory containing pom.xml
		pomFile, err := findBuildFile(repoPath, "java")
		if err != nil {
			return err
		}
		srcDir = filepath.Join(filepath.Dir(pomFile), "target")
	default:
		return fmt.Errorf("unsupported build type for moving artifacts: %s", buildType)
	}

	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("build output directory %s not found", srcDir)
	}
	// Copy all contents of srcDir to outputPath
	cmd := exec.Command("cp", "-r", filepath.Join(srcDir, "*"), outputPath)
	if err := executor.Run(cmd); err != nil {
		return fmt.Errorf("failed to move artifacts from %s to %s: %v", srcDir, outputPath, err)
	}
	logrus.Infof("Moved artifacts from %s to %s", srcDir, outputPath)
	return nil
}
