package provider

import (
	"github.com/eliau2005/ai-git/internal/config"
)

type Provider interface {
	GenerateCommitMessage(diff string, context string) (string, error)
	GetName() string
}

type ProviderFactory struct {
}

func (f *ProviderFactory) GetProvider(name string, pCfg config.ProviderConfig, model string, systemPrompt string, commitPromptTemplate string) Provider {
	switch name {
	case "openai":
		return &OpenAIProvider{
			APIKey:       pCfg.APIKey,
			Model:        model,
			SystemPrompt: systemPrompt,
			CommitPrompt: commitPromptTemplate,
		}
	case "gemini":
		return &GeminiProvider{
			APIKey:       pCfg.APIKey,
			Model:        model,
			SystemPrompt: systemPrompt,
			CommitPrompt: commitPromptTemplate,
		}
	case "ollama":
		return &OllamaProvider{
			BaseURL:      pCfg.BaseURL,
			Model:        model,
			SystemPrompt: systemPrompt,
			CommitPrompt: commitPromptTemplate,
		}
	case "anthropic":
		return &AnthropicProvider{
			APIKey:       pCfg.APIKey,
			Model:        model,
			SystemPrompt: systemPrompt,
			CommitPrompt: commitPromptTemplate,
		}
	default:
		return nil
	}
}
