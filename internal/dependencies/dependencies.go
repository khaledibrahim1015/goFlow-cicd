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

// buildFactory
type buildTypesFactory struct {
	buildType    string
	buildHandler func(
		version string, pm PackageManager) error
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

// SupportedVersions maps build types to valid versions.
var SupportedVersions = map[string][]string{
	dotnet: {"6.0", "7.0", "8.0"},
	java:   {"11", "17", "21"},
}

// // EnsureEnvironment ensures the environment is set up for the given build type and version.
func EnsureEnvironment(buildType, version string) error {

	pm, err := getPackageManager()
	if err != nil {
		return err
	}
	logrus.Infof("Ensuring environment for %s %s using %s on Linux...", buildType, version, pm.Name())

	// Root Validation
	if os.Geteuid() != 0 {
		return fmt.Errorf("root privileges required (run with sudo)")
	}
	// Validate version and build Type
	validVersions, ok := SupportedVersions[buildType]
	if !ok {
		return fmt.Errorf("unsupported build type: %s", buildType)
	}

	if version != "" && !contains(validVersions, version) {
		return fmt.Errorf("invalid version %s for %s; supported: %v", version, buildType, validVersions)
	}
	// so we can easily extend it by add new build types
	for _, buildapp := range AvailableBuildTypes {
		if buildapp.buildType == buildType {
			return buildapp.buildHandler(version, pm)
		}
	}
	return fmt.Errorf("unrecognized build type: %s", buildType)
}

// // ensureDotNet ensures the .NET SDK is installed and configured.
func ensureDotNet(version string, pm PackageManager) error {
	tool := dotnet
	currentVersion, err := getToolVersion(tool, "--version")
	if err == nil && strings.HasPrefix(currentVersion, version) { //dotnet --version -> Expected output 8.0.202
		logrus.Infof(".NET SDK %s found", currentVersion)
		return nil
	}
	logrus.Warnf(".NET SDK not found or wrong version (%s), installing...", currentVersion)

	// Add Microsoft repo if needed (CentOS/RHEL-specific)
	if pm.Name() == "yum" || pm.Name() == "dnf" {
		if _, err := os.Stat("/etc/yum.repos.d/packages-microsoft-prod.repo"); os.IsNotExist(err) {
			repoURL := "https://packages.microsoft.com/config/centos/7/packages-microsoft-prod.rpm"
			if pm.Name() == "dnf" {
				repoURL = "https://packages.microsoft.com/config/centos/8/packages-microsoft-prod.rpm"
			}
			if err := executor.Run(exec.Command(sudo, rpm, "-Uvh", repoURL)); err != nil {
				return fmt.Errorf("failed to add Microsoft .NET repo: %v", err)
			}
		}
	}

	//  pkgs installiation and updates
	pkg := fmt.Sprintf("dotnet-sdk-%s", version)
	if err := pm.Update(); err != nil {
		return err
	}
	if err := pm.Install(pkg); err != nil {
		return fmt.Errorf("failed to install %s: %v", pkg, err)
	}

	// set Environment and configure environment
	dotnetRootPath, err := findDotNetRoot(pm.Name())
	if err != nil {
		return err
	}

	if err := setEnvVar(DOTNET_ROOT, dotnetRootPath); err != nil {
		return err
	}
	if err := appendToPath(filepath.Join(dotnetRootPath, "bin")); err != nil {
		return err
	}

	logrus.Info(".NET SDK installed and environment variables set")
	return VerifyTool(tool)

}

// ensureJava ensures Java and Maven are installed and configured concurrently.
func ensureJava(version string, pm PackageManager) error {
	var wg sync.WaitGroup
	var javaErr, mavenErr error

	// Check and install Java
	wg.Add(1)
	go func() {
		defer wg.Done()
		tool := "java"
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
