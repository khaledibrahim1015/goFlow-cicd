package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// RepositoryConfig defines a Git repository
type RepositoryConfig struct {
	URL    string `json:"url" yaml:"url"`
	Branch string `json:"branch" yaml:"branch"`
	Secret string `json:"secret" yaml:"secret"`
}

// BuildConfig defines the build step
type BuildConfig struct {
	Type string `json:"type" yaml:"type"` // "dotnet", "java" .. etc
	// Path    string `json:"path" yaml:"path"`
	OutputPath string `json:"output_path" yaml:"output_path"` // New field for artifact destination
	Version    string `json:"version" yaml:"version"`         // e.g., "6.0", "11"   .Net sdk version , Java Jdk
}

// TestConfig defines the test step
type TestConfig struct {
	Type string `json:"type" yaml:"type"`
	// Path    string `json:"path" yaml:"path"`
	OutputPath string `json:"output_path" yaml:"output_path"` // New field for artifact destination
	Version    string `json:"version" yaml:"version"`
}
type DeployConfig struct {
	Method string `json:"method" yaml:"method"` // "ssh", "docker", "k8s" , etc
	// Defferent methods configurations
	SSH            *SSHConfig    `json:"ssh,omitempty" yaml:"ssh,omitempty"`
	Docker         *DockerConfig `json:"docker,omitempty" yaml:"docker,omitempty"`
	RollbackScript string        `json:"rollback_script" yaml:"rollback_script"`
	PostDeployCmds []string      `json:"post_deploy_cmds" yaml:"post_deploy_cmds"`
}

// SSHConfig for ssh deployment
type SSHConfig struct {
	RemoteUser   string `json:"remote_user" yaml:"remote_user"`
	RemoteHost   string `json:"remote_host" yaml:"remote_host"`
	RemotePath   string `json:"remote_path" yaml:"remote_path"`
	KeyPath      string `json:"key_path" yaml:"key_path"`
	RsyncOptions string `json:"rsync_options,omitempty"` // Add rsync options field
}

// DockerConfig for Docker deployments
type DockerConfig struct {
	Image       string `json:"image" yaml:"image"`
	Registry    string `json:"registry" yaml:"registry"`
	Username    string `json:"username" yaml:"username"`
	Password    string `json:"password" yaml:"password"`
	ComposeFile string `json:"compose_file" yaml:"compose_file"`
}

// PipelineConfig holds the full configuration
type PipelineConfig struct {
	Repositories []RepositoryConfig `json:"repositories" yaml:"repositories"`
	Build        BuildConfig        `json:"build" yaml:"build"`
	Test         TestConfig         `json:"test" yaml:"test"`
	Deploy       DeployConfig       `json:"deploy" yaml:"deploy"`
}

// func LOadV2
// Enhance config.go to be aligns with industry standards

func Load() (*PipelineConfig, error) {

	// Define Command line flag for config path
	configPath := flag.String("config", "config.json", "path to the configuratioon file ")
	flag.Parse()

	// Check environment variable override (note local environment variable (tem env))
	if envPath := os.Getenv("GOFLOW_CONFIG"); envPath != "" {
		configPath = &envPath
	}

	// If configPath is absolute, use it directly
	if filepath.IsAbs(*configPath) {
		if _, err := os.Stat(*configPath); err == nil {
			logrus.Infof("Loading config from %s (absolute path)", *configPath)
			return loadFromFile(*configPath)
		}
		return nil, fmt.Errorf("config file not found at %s", *configPath)
	}

	execuPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %v", err)
	}

	rootDir := filepath.Dir(execuPath)
	logrus.Infof("Project root directory: %s", rootDir)

	// Possible paths
	possiblePathes := []string{
		filepath.Join(rootDir, *configPath),                       // e.g., /home/khaledibra/projects/config-app/config.json
		filepath.Join(rootDir, "internal", "config", *configPath), // e.g., /home/khaledibra/projects/config-app/internal/configs/config.json
		filepath.Join(os.Getenv("HOME"), ".app", *configPath),     // e.g., /home/khaledibra/.app/config.json
	}
	for _, path := range possiblePathes {
		if _, err := os.Stat(path); err == nil {
			logrus.Infof("Loading config from %s", path)
			return loadFromFile(path)
		}
	}

	return nil, fmt.Errorf("no configuration file found in possible paths: %v", possiblePathes)
	// Get the project root directory (2 levels up from handlers directory)

}

func loadFromFile(path string) (*PipelineConfig, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}

	var cfg PipelineConfig
	switch {
	case strings.HasSuffix(path, ".json"):
		err = json.Unmarshal(data, &cfg)
	case strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml"):
		err = yaml.Unmarshal(data, &cfg)
	default:
		return nil, fmt.Errorf("unsupported format: %s", path)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	// Validation
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %v", err)
	}

	return &cfg, nil
}

func validate(cfg *PipelineConfig) error {
	if len(cfg.Repositories) == 0 {
		return fmt.Errorf("at least one repository required")
	}
	for i, repo := range cfg.Repositories {
		if repo.URL == "" || repo.Branch == "" || repo.Secret == "" {
			return fmt.Errorf("repository %d: url, branch, and secret required", i)
		}
	}

	if cfg.Build.Type == "" {
		return fmt.Errorf("build: type and path required")
	}
	if cfg.Build.OutputPath == "" {
		return fmt.Errorf("build: output_path required")
	}

	if cfg.Build.Type != "dotnet" && cfg.Build.Type != "java" {
		return fmt.Errorf("unsupported build type: %s", cfg.Build.Type)
	}
	if cfg.Test.Type != "dotnet" && cfg.Test.Type != "java" {
		return fmt.Errorf("unsupported test type: %s", cfg.Test.Type)
	}
	if cfg.Deploy.Method != "" {
		switch cfg.Deploy.Method {
		case "ssh":
			if cfg.Deploy.SSH == nil || cfg.Deploy.SSH.RemoteUser == "" || cfg.Deploy.SSH.RemoteHost == "" || cfg.Deploy.SSH.RemotePath == "" {
				return fmt.Errorf("ssh deployment requires remote_user, remote_host, and remote_path")
			}
		case "docker":
			if cfg.Deploy.Docker == nil || cfg.Deploy.Docker.Image == "" {
				return fmt.Errorf("docker deployment requires image")
			}

		default:
			return fmt.Errorf("unsupported deploy method: %s", cfg.Deploy.Method)
		}
	}
	// Validate output_path exists or can be created (optional)
	if err := os.MkdirAll(cfg.Build.OutputPath, 0755); err != nil {
		return fmt.Errorf("invalid output_path %s: %v", cfg.Build.OutputPath, err)
	}
	return nil
}
