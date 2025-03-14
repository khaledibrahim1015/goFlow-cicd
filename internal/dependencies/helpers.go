package dependencies

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	sudo        = "sudo"
	update      = "update"
	install     = "install"
	aptGet      = "apt-get"
	yum         = "yum"
	makecache   = "makecache"
	dnf         = "dnf"
	rpm         = "rpm"
	dotnet      = "dotnet"
	java        = "java"
	DOTNET_ROOT = "DOTNET_ROOT"
)

// Utility Functions
func contains(slice []string, version string) bool {
	for _, val := range slice {
		if val == version {
			return true
		}
	}
	return false
}

func getToolVersion(tool, versionFlag string) (string, error) {
	cmd := exec.Command(tool, versionFlag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s not found or failed: %v", tool, err)
	}
	return strings.TrimSpace(string(output)), nil
}

func findDotNetRoot(pmName string) (string, error) {
	var path string
	switch pmName {
	case aptGet:
		path = "/usr/lib/dotnet"
	case yum, dnf:
		path = "/usr/lib64/dotnet"
	}
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("could not locate .NET SDK root at %s for %s", path, pmName)
}

func setEnvVar(key, value string) error {
	if err := os.Setenv(key, value); err != nil {
		return fmt.Errorf("failed to set %s=%s: %v", key, value, err)
	}
	logrus.Debugf("Set %s=%s", key, value)
	return exportEnvVar(key, value)
}

func exportEnvVar(key, value string) error {
	f, err := os.OpenFile("env_setup.sh", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open env_setup.sh: %v", err)
	}
	defer f.Close()
	if _, err := f.WriteString(fmt.Sprintf("export %s=%s\n", key, value)); err != nil {
		return fmt.Errorf("failed to write %s to env_setup.sh: %v", key, err)
	}
	return nil
}

func appendToPath(newPath string) error {
	currentPath := os.Getenv("PATH")
	if strings.Contains(currentPath, newPath) {
		logrus.Infof("%s already in $PATH", newPath)
		return nil
	}
	path := fmt.Sprintf("%s:%s", currentPath, newPath)
	if err := os.Setenv("PATH", path); err != nil {
		return fmt.Errorf("failed to append %s to PATH: %v", newPath, err)
	}
	logrus.Debugf("Appended %s to PATH", newPath)
	return exportEnvVar("PATH", path)
}
func VerifyTool(tool string) error {
	if _, err := getToolVersion(tool, "--version"); err != nil {
		return fmt.Errorf("%s not usable after installation: %v", tool, err)
	}
	logrus.Infof("%s verified as usable", strings.ToTitle(tool))
	return nil
}

func findJavaHome(pmName, version string) (string, error) {
	var path string
	switch pmName {
	case aptGet:
		path = fmt.Sprintf("/usr/lib/jvm/java-%s-openjdk-amd64", version)
	case yum, dnf:
		path = fmt.Sprintf("/usr/lib/jvm/java-%s-openjdk", version)
	}
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("could not locate JAVA_HOME for version %s at %s for %s", version, path, pmName)
}

func getJavaPackageName(pmName, version string) string {
	switch pmName {
	case aptGet:
		return fmt.Sprintf("openjdk-%s-jdk", version)
	case yum, dnf:
		return fmt.Sprintf("java-%s-openjdk-devel", version)
	default:
		return ""
	}
}

// detects and returns the appropriate package manager
func getPackageManager() (PackageManager, error) {
	for _, pm := range packageManagers {
		if pkgPath, err := exec.LookPath(pm.command); err == nil {
			logrus.Infof("package manager detected: %v", pkgPath)
			return pm.factory(), nil
		}
	}
	return nil, fmt.Errorf("no supported package manager found (apt-get, yum, or dnf required)")
}
