package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/eliau2005/ai-git/internal/config"
	"github.com/eliau2005/ai-git/internal/git"
	"github.com/eliau2005/ai-git/internal/github"
)

func handlePR() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: ai-git pr <create|list>")
		return
	}

	subCmd := os.Args[2]
	switch subCmd {
	case "create":
		handlePRCreate()
	default:
		fmt.Printf("Unknown pr command: %s\n", subCmd)
	}
}

func handlePRCreate() {
	fmt.Println(styleTitle.Render("Create Pull Request"))

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Config error: %v", err)))
		return
	}

	remoteInfo, err := git.GetRemoteInfo()
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Failed to get remote info: %v", err)))
		return
	}

	if remoteInfo.Platform != "github" {
		fmt.Println(styleError.Render("Currently, only GitHub is supported for PR creation."))
		return
	}

	platformCfg, ok := cfg.Platforms[remoteInfo.Platform]
	if !ok || platformCfg.Token == "" {
		fmt.Println(styleError.Render("No token found for GitHub. Please run 'ai-git auth' first."))
		return
	}

	currentBranch, err := git.GetCurrentBranch()
	if err != nil || currentBranch == "" {
		fmt.Println(styleError.Render("Could not determine current branch."))
		return
	}

	var baseBranch string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Base branch").
				Value(&baseBranch).
				Placeholder("main"),
		),
	)

	if err := form.Run(); err != nil {
		return
	}

	if baseBranch == "" {
		baseBranch = "main" // default
	}

	// Diff against base branch
	diff, err := git.DiffBranches(baseBranch, currentBranch)
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Failed to get diff against %s: %v", baseBranch, err)))
		return
	}

	if diff == "" {
		fmt.Println(styleSubtle.Render("No changes found against base branch."))
		return
	}

	// AI Generate PR Content
	// Temporarily repurpose the runAIWorkflow to generate PR description
	contextStr := fmt.Sprintf("Generate a Pull Request Title and Description for these changes. Base: %s, Head: %s.", baseBranch, currentBranch)
	
	finalMsg, ok := runAIWorkflow(diff, contextStr)
	if !ok {
		fmt.Println(styleSubtle.Render("Cancelled."))
		return
	}

	parts := strings.SplitN(finalMsg, "\n", 2)
	title := strings.TrimSpace(parts[0])
	body := ""
	if len(parts) > 1 {
		body = strings.TrimSpace(parts[1])
	}

	var confirm bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Publish PR to GitHub?").
				Value(&confirm),
		),
	)

	if err := confirmForm.Run(); err != nil || !confirm {
		fmt.Println(styleSubtle.Render("Cancelled."))
		return
	}

	// Create PR via GitHub API
	err = runSpinner("Creating PR...", func() error {
		client := github.NewClient(platformCfg.Token)
		_, err := client.CreatePullRequest(context.Background(), remoteInfo.Owner, remoteInfo.Repo, title, body, currentBranch, baseBranch)
		return err
	})

	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Failed to create PR: %v", err)))
	} else {
		fmt.Println(styleSuccess.Render("Pull Request created successfully! 🎉"))
	}
}
