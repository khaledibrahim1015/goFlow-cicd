package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/khaledibrahim1015/goFlow-cicd/internal/server"
	"github.com/khaledibrahim1015/goFlow-cicd/pkg/executor"
	"github.com/sirupsen/logrus"
)

const (
	Github         = "github"
	Gitlab         = "gitlab"
	X_Github_Event = "X-Github-Event"
	X_Gitlab_Event = "X-Gitlab-Event"
)

// that mean the request from githubprovider here we do not need value of X_Github_Event value (push)
// just identify provider
func DetermineGitProvider(req *server.HttpRequest) string {
	if _, err := req.GetHeader(X_Github_Event); err == nil {
		return Github
	}
	if _, err := req.GetHeader(X_Gitlab_Event); err == nil {
		return Gitlab
	}
	return "unknown"
}

// Clone clones a Git repository to a temporary directory and returns its path.
func Clone(url, branch string) (string, error) {
	// Input Validate
	if url == "" {
		return "", fmt.Errorf("repository URL cannot be empty")
	}
	if branch == "" {
		return "", fmt.Errorf("branch cannot be empty")

	}

	if err := validateRepoURL(url); err != nil {
		return "", err
	}
	// Generate unique directory name
	repoName := sanitizedRepoName(url)
	dir, err := os.MkdirTemp("", fmt.Sprintf("goflow-%s-", repoName))
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}

	logrus.Infof("Cloning %s (branch: %s) into %s", url, branch, dir)

	// Prepare and run git clone
	cmd := exec.Command("git", "clone", "--depth", "1", "-b", branch, url, dir)
	output, err := executor.RunWithOutput(cmd)
	if err != nil {
		// Clean up on failure
		if removeErr := os.RemoveAll(dir); removeErr != nil {
			logrus.Warnf("Failed to clean up %s: %v", dir, removeErr)
		}
		return "", fmt.Errorf("clone failed: %v\nOutput: %s", err, output)
	}

	// Verify the directory exists and is a Git repo
	if _, err := os.Stat(filepath.Join(dir, ".git")); os.IsNotExist(err) {
		return "", fmt.Errorf("cloned directory %s is not a valid Git repository", dir)
	}

	logrus.Debugf("Successfully cloned into %s", dir)
	return dir, nil

}

func validateRepoURL(url string) error {
	urlLower := strings.ToLower(url)
	validPrefixes := []string{"http://", "https://", "git@"}
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(urlLower, prefix) {
			return nil
		}
	}
	return fmt.Errorf("invalid repository URL: %s (must start with http://, https://, or git@)", url)
}

// sanitizedRepoName generates a safe, unique name from a Git URL.
func sanitizedRepoName(url string) string {
	// Extract the repo name (e.g., "user-repo" from "https://github.com/user/repo.git")
	parts := strings.Split(strings.TrimSuffix(url, ".git"), string(os.PathSeparator)) //[https:  github.com user repo]
	if len(parts) < 2 {
		return "default-repo"
	}
	return fmt.Sprint("%s-%s", parts[len(parts)-2], parts[len(parts)-1])
}
