package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/ai-git/internal/config"
	"github.com/user/ai-git/internal/git"
	"github.com/user/ai-git/internal/provider"
)

// Global Styles
var (
	styleTitle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).BorderStyle(lipgloss.RoundedBorder()).Padding(0, 1)
	styleSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("#43BF6D"))
	styleError   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87"))
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
		fmt.Println(styleError.Render(fmt.Sprintf("Error: %v", err)))
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
		fmt.Println(styleError.Render(fmt.Sprintf("Error: %v", err)))
		return
	}
	fmt.Println(styleSuccess.Render(fmt.Sprintf("Added %s", os.Args[2])))
}

// --- Spinner Model for AI Generation ---

type spinnerModel struct {
	spinner   spinner.Model
	diff      string
	provider  provider.Provider
	msgResult string
	err       error
	done      bool
}

func initialSpinnerModel(p provider.Provider, diff string) spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return spinnerModel{spinner: s, diff: diff, provider: p}
}

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, generateMsgCmd(m.provider, m.diff))
}

type msgGeneratedMsg struct {
	msg string
	err error
}

func generateMsgCmd(p provider.Provider, diff string) tea.Cmd {
	return func() tea.Msg {
		msg, err := p.GenerateCommitMessage(diff, "")
		return msgGeneratedMsg{msg: msg, err: err}
	}
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case msgGeneratedMsg:
		m.msgResult = msg.msg
		m.err = msg.err
		m.done = true
		return m, tea.Quit
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf("\n %s AI is thinking...\n\n", m.spinner.View())
}

// --- End Spinner Model ---

func handleCommit() {
	fmt.Println(styleTitle.Render("AI Commit"))

	// Step A: Smart Staging
	diff, err := git.DiffStaged()
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Error checking staged changes: %v", err)))
		return
	}

	if diff == "" {
		statusShort, err := git.StatusShort()
		if err != nil {
			fmt.Println(styleError.Render(fmt.Sprintf("Error getting status: %v", err)))
			return
		}

		files := parseGitStatusFiles(statusShort)
		if len(files) == 0 {
			fmt.Println(styleError.Render("No changes to commit."))
			return
		}

		var selectedFiles []string
		
		// Map strings to huh.Options
		var options []huh.Option[string]
		for _, f := range files {
			options = append(options, huh.NewOption(f, f))
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("No staged changes detected. Select files to stage:").
					Options(options...).
					Value(&selectedFiles),
			),
		)

		err = form.Run()
		if err != nil {
			fmt.Println(styleError.Render("Selection cancelled."))
			return
		}

		if len(selectedFiles) == 0 {
			fmt.Println(styleError.Render("No files selected. Aborting."))
			return
		}

		for _, f := range selectedFiles {
			err := git.Add(f)
			if err != nil {
				fmt.Println(styleError.Render(fmt.Sprintf("Error adding %s: %v", f, err)))
				return
			}
		}

		// Re-check diff
		diff, err = git.DiffStaged()
		if err != nil {
			fmt.Println(styleError.Render(fmt.Sprintf("Error checking staged changes: %v", err)))
			return
		}
	}

	// Step B: AI Generation with Feedback
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Error loading config: %v", err)))
		return
	}

	root, err := git.GetRepoRoot()
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Error: %v", err)))
		return
	}

	repoCfg, _ := config.LoadRepoConfig(root)

	selectedProvider := cfg.DefaultProvider
	if repoCfg != nil && repoCfg.EnabledProvider != "" {
		selectedProvider = repoCfg.EnabledProvider
	}

	if selectedProvider == "" {
		fmt.Println(styleError.Render("Error: No AI provider configured. Use 'ai-git config' to set one."))
		return
	}

	pCfg, ok := cfg.Providers[selectedProvider]
	if !ok {
		fmt.Println(styleError.Render(fmt.Sprintf("Error: Provider '%s' not configured.", selectedProvider)))
		return
	}

	model := pCfg.DefaultModel
	if repoCfg != nil && repoCfg.ModelOverride != "" {
		model = repoCfg.ModelOverride
	}

	factory := &provider.ProviderFactory{}
	p := factory.GetProvider(selectedProvider, pCfg, model, cfg.SystemPrompt, cfg.CommitPromptTemplate)
	if p == nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Error: Could not initialize provider '%s'.", selectedProvider)))
		return
	}

	m := initialSpinnerModel(p, diff)
	pProgram := tea.NewProgram(m)
	finalModel, err := pProgram.Run()
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Error running spinner: %v", err)))
		return
	}

	finalSpinnerModel := finalModel.(spinnerModel)
	if finalSpinnerModel.err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("AI Generation Error: %v", finalSpinnerModel.err)))
		return
	}

	generatedMsg := finalSpinnerModel.msgResult

	// Step C: Review & Edit
	parts := strings.SplitN(generatedMsg, "\n", 2)
	title := strings.TrimSpace(parts[0])
	description := ""
	if len(parts) > 1 {
		description = strings.TrimSpace(parts[1])
	}

	for {
		// Summary View
		fmt.Println(lipgloss.NewStyle().MarginTop(1).Render(
			lipgloss.JoinVertical(lipgloss.Left,
				lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Render("Title:"),
				lipgloss.NewStyle().PaddingLeft(2).Render(title),
				"",
				lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Render("Description:"),
				lipgloss.NewStyle().PaddingLeft(2).Render(description),
			),
		))

		var action string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose an action:").
					Options(
						huh.NewOption("Commit Changes", "commit").Selected(true),
						huh.NewOption("Edit Message", "edit"),
						huh.NewOption("Cancel", "cancel"),
					).
					Value(&action),
			),
		)

		err = form.Run()
		if err != nil {
			fmt.Println(styleError.Render("Aborted."))
			return
		}

		if action == "cancel" {
			fmt.Println(styleError.Render("Commit cancelled."))
			return
		}

		if action == "commit" {
			break
		}

		if action == "edit" {
			editForm := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Commit Title").
						Value(&title),
					huh.NewInput(). // Using Input as fallback for multi-line editing
						Title("Commit Description").
						Value(&description),
				),
			)
			err = editForm.Run()
			if err != nil {
				fmt.Println(styleError.Render("Editing cancelled."))
				return
			}
		}
	}

	finalMsg := fmt.Sprintf("%s\n\n%s", title, description)

	// Step D: Execution
	err = git.Commit(finalMsg)
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Commit failed: %v", err)))
		return
	}

	fmt.Println(styleSuccess.Render("Committed successfully."))
}

func handlePush() {
	err := git.Push()
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Error: %v", err)))
		return
	}
	fmt.Println(styleSuccess.Render("Pushed successfully."))
}

func handlePull() {
	err := git.Pull()
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Error: %v", err)))
		return
	}
	fmt.Println(styleSuccess.Render("Pulled successfully."))
}

func handleSync() {
	handleCommit()

	var confirmPush bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Push changes to remote?").
				Value(&confirmPush),
		),
	)

	err := form.Run()
	if err != nil {
		return
	}

	if confirmPush {
		handlePush()
	}
}

func handleInit() {
	root, err := git.GetRepoRoot()
	if err != nil {
		fmt.Println(styleError.Render("Not a git repository."))
		return
	}

	repoConfigPath := filepath.Join(root, ".ai-git.yaml")
	if _, err := os.Stat(repoConfigPath); err == nil {
		fmt.Println(styleSuccess.Render("Repository already initialized with AI-Git."))
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
		fmt.Println(styleError.Render(fmt.Sprintf("Error: %v", err)))
		return
	}
	fmt.Println(styleSuccess.Render("Initialized .ai-git.yaml"))
}

func handleDoctor() {
	fmt.Println(styleTitle.Render("AI-Git Doctor"))

	// Git
	if git.IsRepo() {
		fmt.Println(styleSuccess.Render("[OK] Git repository detected"))
	} else {
		fmt.Println(styleError.Render("[FAIL] Not a git repository"))
	}

	// Config
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("[FAIL] Could not load config: %v", err)))
	} else {
		fmt.Println(styleSuccess.Render("[OK] Configuration loaded"))
		if cfg.DefaultProvider == "" {
			fmt.Println(styleError.Render("[WARN] No default provider set"))
		} else {
			fmt.Printf("[OK] Default provider: %s\n", cfg.DefaultProvider)
			pCfg, ok := cfg.Providers[cfg.DefaultProvider]
			if !ok {
				fmt.Println(styleError.Render(fmt.Sprintf("[FAIL] Provider '%s' configuration missing", cfg.DefaultProvider)))
			} else {
				if pCfg.APIKey == "" && cfg.DefaultProvider != "ollama" {
					fmt.Println(styleError.Render(fmt.Sprintf("[FAIL] API key for '%s' is missing", cfg.DefaultProvider)))
				} else {
					fmt.Println(styleSuccess.Render(fmt.Sprintf("[OK] API key configured for '%s'", cfg.DefaultProvider)))
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
		fmt.Println(styleError.Render(fmt.Sprintf("Error: %v", err)))
		return
	}

	subCommand := os.Args[2]
	switch subCommand {
	case "set-provider":
		if len(os.Args) < 4 {
			fmt.Println(styleError.Render("Missing provider name"))
			return
		}
		cfg.DefaultProvider = os.Args[3]
	case "set-key":
		if len(os.Args) < 5 {
			fmt.Println(styleError.Render("Missing provider name and/or API key"))
			return
		}
		pName := os.Args[3]
		key := os.Args[4]
		pCfg := cfg.Providers[pName]
		pCfg.APIKey = key
		cfg.Providers[pName] = pCfg
	case "set-model":
		if len(os.Args) < 5 {
			fmt.Println(styleError.Render("Missing provider name and/or model name"))
			return
		}
		pName := os.Args[3]
		model := os.Args[4]
		pCfg := cfg.Providers[pName]
		pCfg.DefaultModel = model
		cfg.Providers[pName] = pCfg
	default:
		fmt.Println(styleError.Render(fmt.Sprintf("Unknown config subcommand: %s", subCommand)))
		return
	}

	err = cfg.Save()
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Error saving config: %v", err)))
		return
	}
	fmt.Println(styleSuccess.Render("Configuration updated."))
}

func parseGitStatusFiles(status string) []string {
	var files []string
	lines := strings.Split(status, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Short status format is "XY filename" e.g. "M  file.go" or "?? file.go"
		// XY are the first two chars.
		if len(trimmed) > 3 {
			// Extract filename, handling potential quotes if git does that, though --short usually just lists names.
			// The file name starts after the first 3 characters (status flags + space).
			filePart := strings.TrimSpace(line[2:])
			files = append(files, filePart)
		}
	}
	return files
}