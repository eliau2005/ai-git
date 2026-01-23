package provider

import (
	"github.com/user/ai-git/internal/config"
)

type Provider interface {
	GenerateCommitMessage(diff string, context string) (string, error)
	GetName() string
}

type ProviderFactory struct {
}

func (f *ProviderFactory) GetProvider(name string, pCfg config.ProviderConfig, model string) Provider {
	switch name {
	case "openai":
		return &OpenAIProvider{
			APIKey: pCfg.APIKey,
			Model:  model,
		}
	case "gemini":
		return &GeminiProvider{
			APIKey: pCfg.APIKey,
			Model:  model,
		}
	case "ollama":
		return &OllamaProvider{
			BaseURL: pCfg.BaseURL,
			Model:   model,
		}
	case "anthropic":
		return &AnthropicProvider{
			APIKey: pCfg.APIKey,
			Model:  model,
		}
	default:
		return nil
	}
}
