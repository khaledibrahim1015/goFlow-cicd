package pipeline

import (
	"fmt"
	"os"

	"github.com/khaledibrahim1015/goFlow-cicd/internal/config"
	"github.com/khaledibrahim1015/goFlow-cicd/internal/dependencies"
	"github.com/sirupsen/logrus"
)

type Pipeline struct {
	cfg      *config.PipelineConfig
	repoPath string // which cloned from url that provided
}

func New(cfg *config.PipelineConfig, clonedRepoPath string) *Pipeline {
	return &Pipeline{
		cfg:      cfg,
		repoPath: clonedRepoPath,
	}
}

func (p *Pipeline) Run() error {

	logrus.Info("Starting pipeline...")

	// ephemerial
	defer func() {
		if err := os.RemoveAll(p.repoPath); err != nil {
			logrus.Warnf("Failed to clean up %s: %v", p.repoPath, err)
		} else {
			logrus.Debugf("Cleaned up %s", p.repoPath)
		}
	}()

	if err := dependencies.EnsureEnvironment(p.cfg.Build.Type, p.cfg.Build.Version); err != nil {
		return fmt.Errorf("environment setup failed: %v", err)
	}

	// execute pipeline (build , test , deploy )

	if err := p.build(); err != nil {
		return fmt.Errorf("build failed: %v", err)
	}
	if err := p.test(); err != nil {
		return fmt.Errorf("test failed:%v", err)
	}
	if err := p.deploy(); err != nil {
		return fmt.Errorf("deploy failed:%v", err)
	}
	logrus.Info("Pipeline completed successfully")
	return nil
}
