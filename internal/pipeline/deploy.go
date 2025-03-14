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

func (p *Pipeline) deploy() error {
	if p.cfg.Deploy.Method == "" {
		logrus.Info("No deployment configured, skipping")
		return nil
	}
	logrus.Info("Deploying...")

	switch p.cfg.Deploy.Method {
	case "ssh":
		if err := p.deploySSH(); err != nil {
			logrus.Errorf("SSH deployment failed: %v", err)
			logrus.Info("Executing rollback...")
			if rollbackErr := p.executeRollback(); rollbackErr != nil {
				logrus.Errorf("Rollback failed: %v", rollbackErr)
			}
			return fmt.Errorf("deployment failed: %v", err)
		}
	default:
		return fmt.Errorf("unsupported deploy method: %s", p.cfg.Deploy.Method)
	}
	logrus.Info("Deploy stage completed successfully")
	return nil
}

func (p *Pipeline) deploySSH() error {
	sshConfig := p.cfg.Deploy.SSH
	if sshConfig == nil {
		return fmt.Errorf("SSH config missing")
	}

	// Validate required SSH configuration fields
	if sshConfig.RemoteUser == "" || sshConfig.RemoteHost == "" || sshConfig.RemotePath == "" || sshConfig.KeyPath == "" {
		return fmt.Errorf("incomplete SSH configuration")
	}

	// Ensure artifacts exist
	srcDir := p.cfg.Build.OutputPath
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("build artifacts not found at %s", srcDir)
	}

	// Construct rsync command with SSH
	rsyncCmd := []string{
		"rsync",
		"-e", fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no", sshConfig.KeyPath), // Use SSH with the key
		"-avz", // Archive mode, verbose, compress
	}
	logrus.Infof("COMMAND : %v", strings.Join(rsyncCmd, " "))

	// Add custom rsync options if specified
	if sshConfig.RsyncOptions != "" {
		rsyncCmd = append(rsyncCmd, strings.Split(sshConfig.RsyncOptions, " ")...)
	}

	// Add source and destination paths
	sourcePath := srcDir
	if !strings.HasSuffix(sourcePath, "/") {
		sourcePath += "/"
	}
	rsyncCmd = append(rsyncCmd,
		sourcePath, // Source directory (build output)
		fmt.Sprintf("%s@%s:%s", sshConfig.RemoteUser, sshConfig.RemoteHost, sshConfig.RemotePath), // Destination
	)

	// Execute rsync command
	cmd := exec.Command(rsyncCmd[0], rsyncCmd[1:]...)
	cmd.Dir = filepath.Dir(sourcePath)

	output, err := executor.RunWithOutput(cmd)
	if err != nil {
		logrus.Errorf("Deploy failed: %v\nOutput: %s", err, output)
		return fmt.Errorf("rsync failed: %v", err)
	}

	logrus.Info("Deploy succeeded")
	logrus.Debugf("Deploy output: %s", output)

	// Execute post-deployment commands (including running the application)
	if len(p.cfg.Deploy.PostDeployCmds) > 0 {
		for _, postCmd := range p.cfg.Deploy.PostDeployCmds {
			sshCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no %s@%s %s", sshConfig.KeyPath, sshConfig.RemoteUser, sshConfig.RemoteHost, postCmd)
			cmd = exec.Command("sh", "-c", sshCmd)
			output, err := executor.RunWithOutput(cmd)
			if err != nil {
				logrus.Errorf("Post-deploy command '%s' failed: %v\nOutput: %s", postCmd, err, output)
				return fmt.Errorf("post-deploy command '%s' failed: %v", postCmd, err)
			}
			logrus.Infof("Executed post-deploy command: %s", postCmd)
			logrus.Debugf("Post-deploy output: %s", output)
		}
	}

	logrus.Info("SSH deployment successful")
	return nil
}

func (p *Pipeline) executeRollback() error {
	sshConfig := p.cfg.Deploy.SSH
	if sshConfig == nil {
		return fmt.Errorf("SSH config missing for rollback")
	}

	// If a rollback script is specified, execute it locally
	if p.cfg.Deploy.RollbackScript != "" {
		cmd := exec.Command("bash", p.cfg.Deploy.RollbackScript)
		cmd.Env = os.Environ()
		cmd.Dir = p.repoPath
		output, err := executor.RunWithOutput(cmd)
		if err != nil {
			return fmt.Errorf("rollback script failed: %v\nOutput: %s", err, output)
		}
		logrus.Info("Rollback script executed successfully")
		logrus.Debugf("Rollback output: %s", output)
		return nil
	}

	// Default rollback: remove deployed files on the remote server
	rollbackCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no %s@%s 'rm -rf %s/*'", sshConfig.KeyPath, sshConfig.RemoteUser, sshConfig.RemoteHost, sshConfig.RemotePath)
	cmd := exec.Command("sh", "-c", rollbackCmd)
	output, err := executor.RunWithOutput(cmd)
	if err != nil {
		return fmt.Errorf("remote file removal failed: %v\nOutput: %s", err, output)
	}
	logrus.Info("Remote files removed successfully during rollback")
	logrus.Debugf("Rollback output: %s", output)
	return nil
}
