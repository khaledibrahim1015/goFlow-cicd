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
	Name() string // For logging purposes

}

// packageManagerDef defines a package manager to check
type packageManagerDef struct {
	command string
	factory func() PackageManager
}

// Available package managers
var packageManagers = []packageManagerDef{
	{
		command: apt_get,
		factory: func() PackageManager { return AptGetManager{} },
	},
	{
		command: yum,
		factory: func() PackageManager { return YumManager{} },
	},
	{
		command: dnf,
		factory: func() PackageManager { return DnfManager{} },
	},
}

// getPackageManager detects and returns the appropriate package manager
func getPackageManager() (PackageManager, error) {
	for _, pm := range packageManagers {
		if pkgPath, err := exec.LookPath(pm.command); err == nil {
			logrus.Infof("package manager detected: %v", pkgPath)
			return pm.factory(), nil
		}
	}
	return nil, fmt.Errorf("no supported package manager found (apt-get, yum, or dnf required)")
}

// AptGetManager impelements PackageManager for Debian/Ubuntu.
type AptGetManager struct{}

func (apt AptGetManager) Update() error {
	return executor.Run(exec.Command(sudo, apt_get, update))
}
func (apt AptGetManager) Install(pkg string) error {
	return executor.Run(exec.Command(sudo, apt_get, install, "-y", pkg))
}
func (apt AptGetManager) Name() string {
	return "apt-get"
}

// YumManager implements PackageManager for CentOS/RHEL (older versions).
type YumManager struct{}

func (y YumManager) Update() error {
	return executor.Run(exec.Command(sudo, yum, makecache))
}
func (y YumManager) Install(pkg string) error {
	return executor.Run(exec.Command(sudo, yum, install, "-y", pkg))
}
func (y YumManager) Name() string {
	return "yum"
}

// DnfManager implements PackageManager for CentOS 8+/RHEL 8+.
type DnfManager struct{}

func (m DnfManager) Update() error {
	return executor.Run(exec.Command(sudo, dnf, makecache))
}

func (m DnfManager) Install(pkg string) error {
	return executor.Run(exec.Command(sudo, dnf, install, "-y", pkg))
}

func (m DnfManager) Name() string {
	return "dnf"
}
