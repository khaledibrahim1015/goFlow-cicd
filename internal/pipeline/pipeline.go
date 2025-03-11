package pipeline

import (
	"fmt"

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
	if err := dependencies.EnsureEnvironment(p.cfg.Build.Type, p.cfg.Build.Version); err != nil {
		return fmt.Errorf("environment setup failed: %v", err)
	}

	// execute pipeline (build , test , deploy )

	if err := p.build(); err != nil {
		return fmt.Errorf("build failed: %v", err)
	}
	// if err := p.test(); err != nil {
	// 	return fmt.Errorf("test failed:%v", err)
	// }
	// if err := p.deploy(); err != nil {
	// 	// here impelement scripts
	// 	return err
	// }
	logrus.Info("Pipeline completed successfully")
	return nil
}

//     if err := p.deploy(); err != nil {
//         if p.cfg.Deploy.RollbackScript != "" {
//             logrus.Warn("Deployment failed, attempting rollback...")
//             cmd := exec.Command("sh", "-c", p.cfg.Deploy.RollbackScript)
//             if err := executor.Run(cmd); err != nil {
//                 logrus.Errorf("Rollback failed: %v", err)
//             } else {
//                 logrus.Info("Rollback successful")
//             }
//         }
//         return fmt.Errorf("deploy failed: %v", err)
//     }

//     logrus.Info("Pipeline completed successfully")
//     return nil
// }
