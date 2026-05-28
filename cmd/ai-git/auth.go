package main

import (
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/eliau2005/ai-git/internal/config"
)

func handleAuth() {
	fmt.Println(styleTitle.Render("Authentication"))

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Config Error: %v", err)))
		return
	}

	var platform string
	var token string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Platform").
				Options(
					huh.NewOption("GitHub", "github"),
					huh.NewOption("GitLab", "gitlab"),
				).
				Value(&platform),
			huh.NewInput().
				Title("Personal Access Token (PAT)").
				Password(true).
				Value(&token),
		),
	)

	if err := form.Run(); err != nil {
		return
	}

	if cfg.Platforms == nil {
		cfg.Platforms = make(map[string]config.PlatformConfig)
	}

	cfg.Platforms[platform] = config.PlatformConfig{
		Token: token,
	}

	if err := cfg.Save(); err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Error saving auth: %v", err)))
	} else {
		fmt.Println(styleSuccess.Render(fmt.Sprintf("Successfully authenticated with %s", platform)))
	}
}
