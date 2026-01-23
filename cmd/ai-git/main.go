package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/ai-git/internal/config"
	"github.com/user/ai-git/internal/git"
	"github.com/user/ai-git/internal/provider"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "status":
		handleStatus()
	case "add":
		handleAdd()
	case "commit":
		handleCommit()
	case "push":
		handlePush()
	case "pull":
		handlePull()
	case "sync":
		handleSync()
	case "init":
		handleInit()
	case "config":
		handleConfig()
	case "doctor":
		handleDoctor()
	case "version":
		fmt.Println("ai-git version 0.1.0")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: ai-git <command> [args]")
	fmt.Println("Commands:")
	fmt.Println("  init    Initialize repository as AI-Git enabled")
	fmt.Println("  status  Show repository status")
	fmt.Println("  add     Stage changes")
	fmt.Println("  commit  Create commit with AI-generated message")
	fmt.Println("  push    Push commits to remote")
	fmt.Println("  pull    Fetch and merge remote changes")
	fmt.Println("  sync    Combined status -> add -> commit -> push")
	fmt.Println("  config  Manage configuration")
	fmt.Println("  doctor  Validate setup")
	fmt.Println("  version Show version info")
}

func handleStatus() {
	out, err := git.Status()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Print(out)
}

func handleAdd() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: ai-git add <path>")
		return
	}
	err := git.Add(os.Args[2])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Added", os.Args[2])
}

func handleCommit() {
	// 1. Show Status
	fmt.Println("--- Git Status ---")
	out, err := git.Status()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Print(out)
	fmt.Println("------------------")

	// 2. Stage Changes (Interactive-ish)
	// For "smart" behavior, we'll try to add everything if nothing is staged,
	// but let's just follow the user's "status, add" flow request.
	// We will ask to stage all changes.
	fmt.Print("Stage all changes? (y/n): ")
	var confirmAdd string
	fmt.Scanln(&confirmAdd)
	if confirmAdd == "y" || confirmAdd == "Y" {
		err := git.Add(".")
		if err != nil {
			fmt.Printf("Error adding files: %v\n", err)
			return
		}
	}

	// 3. Show Status Again
	fmt.Println("--- Updated Status ---")
	out, err = git.Status()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Print(out)
	fmt.Println("----------------------")

	// 4. Proceed with Generation
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	root, err := git.GetRepoRoot()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	repoCfg, _ := config.LoadRepoConfig(root)

	selectedProvider := cfg.DefaultProvider
	if repoCfg != nil && repoCfg.EnabledProvider != "" {
		selectedProvider = repoCfg.EnabledProvider
	}

	if selectedProvider == "" {
		fmt.Println("Error: No AI provider configured. Use 'ai-git config' to set one.")
		return
	}

	pCfg, ok := cfg.Providers[selectedProvider]
	if !ok {
		fmt.Printf("Error: Provider '%s' not configured.\n", selectedProvider)
		return
	}

	model := pCfg.DefaultModel
	if repoCfg != nil && repoCfg.ModelOverride != "" {
		model = repoCfg.ModelOverride
	}

	diff, err := git.DiffStaged()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if diff == "" {
		fmt.Println("No staged changes to commit.")
		return
	}

	factory := &provider.ProviderFactory{}
	p := factory.GetProvider(selectedProvider, pCfg, model)
	if p == nil {
		fmt.Printf("Error: Could not initialize provider '%s'.\n", selectedProvider)
		return
	}

	fmt.Println("Generating commit message...")
	msg, err := p.GenerateCommitMessage(diff, "")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Proposed commit message:\n---\n%s\n---\n", msg)
	fmt.Print("Confirm commit? (y/n): ")
	var confirm string
	fmt.Scanln(&confirm)

	if confirm == "y" || confirm == "Y" {
		err := git.Commit(msg)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Println("Committed successfully.")
	} else {
		fmt.Println("Commit cancelled.")
	}
}

func handlePush() {
	err := git.Push()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Pushed successfully.")
}

func handlePull() {
	err := git.Pull()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Pulled successfully.")
}

func handleSync() {
	// handleCommit now handles status and adding
	handleCommit()

	fmt.Print("Push changes? (y/n): ")
	var confirmPush string
	fmt.Scanln(&confirmPush)
	if confirmPush == "y" || confirmPush == "Y" {
		handlePush()
	}
}

func handleInit() {
	root, err := git.GetRepoRoot()
	if err != nil {
		fmt.Println("Not a git repository.")
		return
	}

	repoConfigPath := filepath.Join(root, ".ai-git.yaml")
	if _, err := os.Stat(repoConfigPath); err == nil {
		fmt.Println("Repository already initialized with AI-Git.")
		return
	}

	cfg, _ := config.LoadConfig()
	defaultProvider := "openai"
	if cfg != nil && cfg.DefaultProvider != "" {
		defaultProvider = cfg.DefaultProvider
	}

	content := fmt.Sprintf("enabled_provider: %s\ncommit_style: conventional\nlanguage: english\n", defaultProvider)
	err = os.WriteFile(repoConfigPath, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Initialized .ai-git.yaml")
}

func handleDoctor() {
	fmt.Println("Checking AI-Git setup...")

	// Git
	if git.IsRepo() {
		fmt.Println("[OK] Git repository detected")
	} else {
		fmt.Println("[FAIL] Not a git repository")
	}

	// Config
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("[FAIL] Could not load config: %v\n", err)
	} else {
		fmt.Println("[OK] Configuration loaded")
		if cfg.DefaultProvider == "" {
			fmt.Println("[WARN] No default provider set")
		} else {
			fmt.Printf("[OK] Default provider: %s\n", cfg.DefaultProvider)
			pCfg, ok := cfg.Providers[cfg.DefaultProvider]
			if !ok {
				fmt.Printf("[FAIL] Provider '%s' configuration missing\n", cfg.DefaultProvider)
			} else {
				if pCfg.APIKey == "" && cfg.DefaultProvider != "ollama" {
					fmt.Printf("[FAIL] API key for '%s' is missing\n", cfg.DefaultProvider)
				} else {
					fmt.Printf("[OK] API key configured for '%s'\n", cfg.DefaultProvider)
				}
			}
		}
	}
}

func handleConfig() {
	if len(os.Args) < 3 {
		fmt.Println("Usage:")
		fmt.Println("  ai-git config set-provider <provider>")
		fmt.Println("  ai-git config set-key <provider> <api-key>")
		fmt.Println("  ai-git config set-model <provider> <model>")
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	subCommand := os.Args[2]
	switch subCommand {
	case "set-provider":
		if len(os.Args) < 4 {
			fmt.Println("Missing provider name")
			return
		}
		cfg.DefaultProvider = os.Args[3]
	case "set-key":
		if len(os.Args) < 5 {
			fmt.Println("Missing provider name and/or API key")
			return
		}
		pName := os.Args[3]
		key := os.Args[4]
		pCfg := cfg.Providers[pName]
		pCfg.APIKey = key
		cfg.Providers[pName] = pCfg
	case "set-model":
		if len(os.Args) < 5 {
			fmt.Println("Missing provider name and/or model name")
			return
		}
		pName := os.Args[3]
		model := os.Args[4]
		pCfg := cfg.Providers[pName]
		pCfg.DefaultModel = model
		cfg.Providers[pName] = pCfg
	default:
		fmt.Printf("Unknown config subcommand: %s\n", subCommand)
		return
	}

	err = cfg.Save()
	if err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		return
	}
	fmt.Println("Configuration updated.")
}
