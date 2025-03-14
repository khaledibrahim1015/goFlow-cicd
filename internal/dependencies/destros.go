package dependencies

import (
	"fmt"
	"os/exec"

	"github.com/khaledibrahim1015/goFlow-cicd/pkg/executor"
	"github.com/sirupsen/logrus"
)

// PackageManager defines an interface for package managment for different distros
type PackageManager interface {
	Update() error
	Install(pkg string) error
	Name() string
}

// packageManagerDef defines a package manager to check
type packageManagerDef struct {
	command string
	factory func() PackageManager
}

// Available package managers
var packageManagers = []packageManagerDef{
	{command: aptGet, factory: func() PackageManager { return AptGetManager{} }},
	{command: yum, factory: func() PackageManager { return YumManager{} }},
	{command: dnf, factory: func() PackageManager { return DnfManager{} }},
}

// AptGetManager implements PackageManager for Debian/Ubuntu
type AptGetManager struct{}

func (apt AptGetManager) Name() string { return aptGet }
func (apt AptGetManager) Update() error {
	cmd := exec.Command(sudo, aptGet, update)
	output, err := executor.RunWithOutput(cmd)
	if err != nil {
		return fmt.Errorf("%s update failed: %v\nOutput: %s", aptGet, err, output)
	}
	logrus.Infof("Output: %s", output)
	return nil
}
func (apt AptGetManager) Install(pkg string) error {
	cmd := exec.Command(sudo, aptGet, install, "-y", pkg)
	output, err := executor.RunWithOutput(cmd)
	if err != nil {
		return fmt.Errorf("%s install %s failed: %v\nOutput: %s", aptGet, pkg, err, output)
	}
	logrus.Infof("Output: %s", output)
	return nil
}

// YumManager implements PackageManager for CentOS 7
type YumManager struct{}

func (y YumManager) Name() string { return yum }
func (y YumManager) Update() error {
	cmd := exec.Command(sudo, yum, update, "-y")
	output, err := executor.RunWithOutput(cmd)
	if err != nil {
		return fmt.Errorf("%s update failed: %v\nOutput: %s", yum, err, output)
	}
	logrus.Infof("Output: %s", output)
	return nil
}
func (y YumManager) Install(pkg string) error {
	cmd := exec.Command(sudo, yum, install, "-y", pkg)
	output, err := executor.RunWithOutput(cmd)
	if err != nil {
		return fmt.Errorf("%s install %s failed: %v\nOutput: %s", yum, pkg, err, output)
	}
	logrus.Infof("Output: %s", output)
	return nil
}

// DnfManager implements PackageManager for CentOS 8/9
type DnfManager struct{}

func (d DnfManager) Name() string { return dnf }
func (d DnfManager) Update() error {
	cmd := exec.Command(sudo, dnf, makecache)
	output, err := executor.RunWithOutput(cmd)
	if err != nil {
		return fmt.Errorf("%s makecache failed: %v\nOutput: %s", dnf, err, output)
	}
	logrus.Infof("Output: %s", output)
	return nil
}
func (d DnfManager) Install(pkg string) error {
	cmd := exec.Command(sudo, dnf, install, "-y", pkg)
	output, err := executor.RunWithOutput(cmd)
	if err != nil {
		return fmt.Errorf("%s install %s failed: %v\nOutput: %s", dnf, pkg, err, output)
	}
	logrus.Infof("Output: %s", output)
	return nil
}
