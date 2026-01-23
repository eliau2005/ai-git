package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultProvider      string                    `yaml:"default_provider"`
	Providers            map[string]ProviderConfig `yaml:"providers"`
	Output               OutputConfig              `yaml:"output"`
	SystemPrompt         string                    `yaml:"system_prompt,omitempty"`
	CommitPromptTemplate string                    `yaml:"commit_prompt_template,omitempty"`
}

type ProviderConfig struct {
	APIKey       string            `yaml:"api_key"`
	DefaultModel string            `yaml:"default_model"`
	CustomModels []string          `yaml:"custom_models,omitempty"`
	BaseURL      string            `yaml:"base_url,omitempty"`
}

type OutputConfig struct {
	Language string `yaml:"language"`
	Style    string `yaml:"style"`
}

type RepoConfig struct {
	EnabledProvider string `yaml:"enabled_provider"`
	ModelOverride   string `yaml:"model_override"`
	CommitStyle     string `yaml:"commit_style"`
	Language        string `yaml:"language"`
}

func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(home, ".config", "ai-git", "config.yaml")

	defaultSystemPrompt := "You are an expert developer. Generate a raw git commit message. Output ONLY the message. Structure: a short title, then a blank line, then a description. No conversational filler, no quotes, no backticks."
	defaultCommitPrompt := "Generate a raw git commit message for the changes below. Output ONLY the message. Structure: a short title, then a blank line, then a description. No conversational filler, no quotes, no backticks.\n\nChanges:\n%s\n\n%s"

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Providers:            make(map[string]ProviderConfig),
				SystemPrompt:         defaultSystemPrompt,
				CommitPromptTemplate: defaultCommitPrompt,
			}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.SystemPrompt == "" {
		cfg.SystemPrompt = defaultSystemPrompt
	}
	if cfg.CommitPromptTemplate == "" {
		cfg.CommitPromptTemplate = defaultCommitPrompt
	}

	return &cfg, nil
}

func LoadRepoConfig(rootPath string) (*RepoConfig, error) {
	repoConfigPath := filepath.Join(rootPath, ".ai-git.yaml")
	data, err := os.ReadFile(repoConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var cfg RepoConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (cfg *Config) Save() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(home, ".config", "ai-git")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "config.yaml")

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}