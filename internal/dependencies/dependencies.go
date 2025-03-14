package dependencies

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/khaledibrahim1015/goFlow-cicd/pkg/executor"
	"github.com/sirupsen/logrus"
)

// buildTypesFactory defines a build type and its handler

type buildTypesFactory struct {
	buildType    string
	buildHandler func(
		version string, pm PackageManager) error
}

// SupportedVersions maps build types to valid versions
var SupportedVersions = map[string][]string{
	dotnet: {"6.0", "7.0", "8.0", "9.0"},
	java:   {"11", "17", "21"},
}

// Available Build Types
var AvailableBuildTypes = []buildTypesFactory{
	{
		buildType:    dotnet,
		buildHandler: ensureDotNet,
	},
	{
		buildType:    java,
		buildHandler: ensureJava,
	},
}

// EnsureEnvironment ensures the environment is set up for the given build type and version
func EnsureEnvironment(buildType, version string) error {
	pm, err := getPackageManager()
	if err != nil {
		return err
	}
	logrus.Infof("package manager detected: /usr/bin/%s", pm.Name())
	logrus.Infof("Ensuring environment for %s %s using %s on Linux...", buildType, version, pm.Name())

	if os.Geteuid() != 0 {
		return fmt.Errorf("root privileges required (run with sudo)")
	}

	validVersions, ok := SupportedVersions[buildType]
	if !ok {
		return fmt.Errorf("unsupported build type: %s", buildType)
	}
	if version != "" && !contains(validVersions, version) {
		return fmt.Errorf("invalid version %s for %s; supported: %v", version, buildType, validVersions)
	}

	for _, buildapp := range AvailableBuildTypes {
		if buildapp.buildType == buildType {
			return buildapp.buildHandler(version, pm)
		}
	}
	return fmt.Errorf("unrecognized build type: %s", buildType)
}

// ensureDotNet ensures the .NET SDK is installed and configured
func ensureDotNet(version string, pm PackageManager) error {
	tool := dotnet
	currentVersion, err := getToolVersion(tool, "--version")
	if err == nil && strings.HasPrefix(currentVersion, version) {
		logrus.Infof(".NET SDK %s found", currentVersion)
		dotnetRoot := os.Getenv(DOTNET_ROOT)
		if dotnetRoot == "" {
			dotnetRoot, err = findDotNetRoot(pm.Name())
			if err != nil {
				return err
			}
			if err := setEnvVar(DOTNET_ROOT, dotnetRoot); err != nil {
				return err
			}
		}
		return appendToPath(filepath.Join(dotnetRoot, "bin"))
	}
	logrus.Warnf(".NET SDK not found or wrong version (%s), installing...", currentVersion)

	// Add Microsoft repo
	switch pm.Name() {
	case aptGet:
		if _, err := os.Stat("/etc/apt/sources.list.d/microsoft-prod.list"); os.IsNotExist(err) {
			cmd := exec.Command(sudo, "wget", "https://packages.microsoft.com/config/ubuntu/24.04/packages-microsoft-prod.deb", "-O", "/tmp/packages-microsoft-prod.deb")
			if output, err := executor.RunWithOutput(cmd); err != nil {
				return fmt.Errorf("failed to fetch Microsoft repo for %s: %v\nOutput: %s", pm.Name(), err, output)
			}
			cmd = exec.Command(sudo, "dpkg", "-i", "/tmp/packages-microsoft-prod.deb")
			if output, err := executor.RunWithOutput(cmd); err != nil {
				return fmt.Errorf("failed to install Microsoft repo for %s: %v\nOutput: %s", pm.Name(), err, output)
			}
		}
	case yum:
		if _, err := os.Stat("/etc/yum.repos.d/microsoft-prod.repo"); os.IsNotExist(err) {
			cmd := exec.Command(sudo, rpm, "-Uvh", "https://packages.microsoft.com/config/centos/7/packages-microsoft-prod.rpm")
			if output, err := executor.RunWithOutput(cmd); err != nil {
				return fmt.Errorf("failed to add Microsoft .NET repo for %s: %v\nOutput: %s", yum, err, output)
			}
		}
	case dnf:
		if _, err := os.Stat("/etc/yum.repos.d/microsoft-prod.repo"); os.IsNotExist(err) {
			cmd := exec.Command(sudo, rpm, "-Uvh", "https://packages.microsoft.com/config/centos/8/packages-microsoft-prod.rpm")
			if output, err := executor.RunWithOutput(cmd); err != nil {
				return fmt.Errorf("failed to add Microsoft .NET repo for %s: %v\nOutput: %s", dnf, err, output)
			}
		}
	}

	pkg := fmt.Sprintf("dotnet-sdk-%s", version)
	if err := pm.Update(); err != nil {
		return err
	}
	if err := pm.Install(pkg); err != nil {
		return fmt.Errorf("failed to install %s: %v", pkg, err)
	}

	dotnetRoot, err := findDotNetRoot(pm.Name())
	if err != nil {
		cmd := exec.Command("which", "dotnet")
		if output, err := cmd.Output(); err == nil {
			dotnetPath := strings.TrimSpace(string(output))
			dotnetRoot = filepath.Dir(dotnetPath)
			if _, err := os.Stat(dotnetRoot); err != nil {
				return err
			}
			logrus.Infof("Inferred .NET SDK root from PATH: %s", dotnetRoot)
		} else {
			return err
		}
	}

	if err := setEnvVar(DOTNET_ROOT, dotnetRoot); err != nil {
		return err
	}
	if err := appendToPath(filepath.Join(dotnetRoot, "bin")); err != nil {
		return err
	}

	logrus.Info(".NET SDK installed and environment variables set")
	return VerifyTool(tool)
}

// ensureJava ensures Java and Maven are installed and configured concurrently
func ensureJava(version string, pm PackageManager) error {
	var wg sync.WaitGroup
	var javaErr, mavenErr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		tool := java
		currentVersion, err := getToolVersion(tool, "-version")
		if err == nil && strings.Contains(currentVersion, version) {
			logrus.Infof("Java %s found", currentVersion)
			return
		}
		logrus.Warnf("Java not found or wrong version (%s), installing...", currentVersion)

		pkg := getJavaPackageName(pm.Name(), version)
		if err := pm.Update(); err != nil {
			javaErr = err
			return
		}
		if err := pm.Install(pkg); err != nil {
			javaErr = fmt.Errorf("failed to install %s: %v", pkg, err)
			return
		}

		javaHome, err := findJavaHome(pm.Name(), version)
		if err != nil {
			javaErr = err
			return
		}
		if err := setEnvVar("JAVA_HOME", javaHome); err != nil {
			javaErr = err
			return
		}
		if err := appendToPath(filepath.Join(javaHome, "bin")); err != nil {
			javaErr = err
			return
		}
		logrus.Info("Java installed and environment variables set")
	}()

	// Check and install Maven
	wg.Add(1)
	go func() {
		defer wg.Done()
		tool := "mvn"
		if _, err := getToolVersion(tool, "--version"); err == nil {
			logrus.Info("Maven found")
			return
		}
		logrus.Warn("Maven not found, installing...")

		if err := pm.Install("maven"); err != nil {
			mavenErr = fmt.Errorf("failed to install maven: %v", err)
			return
		}
		logrus.Info("Maven installed")
	}()

	wg.Wait()
	if javaErr != nil {
		return javaErr
	}
	if mavenErr != nil {
		return mavenErr
	}

	return VerifyTool("mvn")
}
