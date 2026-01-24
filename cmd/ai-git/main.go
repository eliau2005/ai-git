package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	styleSubtle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
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
		fmt.Println("ai-git version 0.2.0")
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
	fmt.Println("  add     Stage changes (run without args for interactive mode)")
	fmt.Println("  commit  Create commit with AI-generated message")
	fmt.Println("  push    Push commits to remote")
	fmt.Println("  pull    Fetch and merge remote changes")
	fmt.Println("  sync    Combined status -> add -> commit -> push")
	fmt.Println("  config  Manage configuration (run without args for interactive mode)")
	fmt.Println("  doctor  Validate setup")
	fmt.Println("  version Show version info")
}

// --- Generic Spinner Helper ---

type actionSpinnerModel struct {
	spinner spinner.Model
	action  func() error
	err     error
	done      bool
	title   string
}

func newActionSpinner(title string, action func() error) actionSpinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return actionSpinnerModel{spinner: s, action: action, title: title}
}

func (m actionSpinnerModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, func() tea.Msg {
		return m.action()
	})
}

func (m actionSpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case error:
		m.err = msg
		m.done = true
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	// Check if the action returned nil (success)
	if msg == nil {
		m.done = true
		return m, tea.Quit
	}
	return m, nil
}

func (m actionSpinnerModel) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf(" %s %s", m.spinner.View(), m.title)
}

func runSpinner(title string, action func() error) error {
	m := newActionSpinner(title, action)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return err
	}
	finalState := finalModel.(actionSpinnerModel)
	return finalState.err
}

// --- Commands ---

func handleStatus() {
	out, err := git.Status()
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Error: %v", err)))
		return
	}
	fmt.Print(out)
}

func handleAdd() {
	// If args provided, use standard git add
	if len(os.Args) >= 3 {
		err := git.Add(os.Args[2])
		if err != nil {
			fmt.Println(styleError.Render(fmt.Sprintf("Error: %v", err)))
			return
		}
		fmt.Println(styleSuccess.Render(fmt.Sprintf("Added %s", os.Args[2])))
		return
	}

	// Interactive Mode
	fmt.Println(styleTitle.Render("Interactive Stage"))
	statusShort, err := git.StatusShort()
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Error getting status: %v", err)))
		return
	}

	files := parseGitStatusFiles(statusShort)
	if len(files) == 0 {
		fmt.Println(styleSuccess.Render("No changed files to stage."))
		return
	}

	var selectedFiles []string
	var options []huh.Option[string]
	for _, f := range files {
		options = append(options, huh.NewOption(f, f))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select files to stage:").
				Options(options...).
				Value(&selectedFiles),
		),
	)

	err = form.Run()
	if err != nil {
		return
	}

	if len(selectedFiles) == 0 {
		fmt.Println("No files selected.")
		return
	}

	err = runSpinner("Staging files...", func() error {
		for _, f := range selectedFiles {
			if err := git.Add(f); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Error staging files: %v", err)))
	} else {
		fmt.Println(styleSuccess.Render("Files staged successfully."))
	}
}

// --- Spinner for AI (Specific) ---


type aiSpinnerModel struct {
	spinner   spinner.Model
	diff      string
	provider  provider.Provider
	msgResult string
	err       error
	done      bool
}

func initialAISpinner(p provider.Provider, diff string) aiSpinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return aiSpinnerModel{spinner: s, diff: diff, provider: p}
}

func (m aiSpinnerModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, func() tea.Msg {
		msg, err := m.provider.GenerateCommitMessage(m.diff, "")
		return msgGeneratedMsg{msg: msg, err: err}
	})
}

type msgGeneratedMsg struct {
	msg string
	err error
}

func (m aiSpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m aiSpinnerModel) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf("\n %s AI is thinking...\n\n", m.spinner.View())
}

func handleCommit() {
	fmt.Println(styleTitle.Render("AI Commit"))

	// Smart Staging Check
	diff, err := git.DiffStaged()
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Error checking staged changes: %v", err)))
		return
	}

	if diff == "" {
		// reuse interactive add logic inline? Or just call handleAdd()? 
		// handleAdd() with no args does exactly what we want, but we need to know if it succeeded.
		// Let's replicate the logic briefly to control flow.
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

		if err := form.Run(); err != nil {
			return
		}
		if len(selectedFiles) == 0 {
			fmt.Println(styleError.Render("Aborted."))
			return
		}

		for _, f := range selectedFiles {
			git.Add(f)
		}
		
		diff, _ = git.DiffStaged()
	}

	// AI Gen
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Config Error: %v", err)))
		return
	}

	root, _ := git.GetRepoRoot()
	repoCfg, _ := config.LoadRepoConfig(root)

	selectedProvider := cfg.DefaultProvider
	if repoCfg != nil && repoCfg.EnabledProvider != "" {
		selectedProvider = repoCfg.EnabledProvider
	}
	if selectedProvider == "" {
		fmt.Println(styleError.Render("No AI provider configured."))
		return
	}

	pCfg, ok := cfg.Providers[selectedProvider]
	if !ok {
		fmt.Println(styleError.Render("Provider not configured."))
		return
	}

	model := pCfg.DefaultModel
	if repoCfg != nil && repoCfg.ModelOverride != "" {
		model = repoCfg.ModelOverride
	}

	factory := &provider.ProviderFactory{}
	p := factory.GetProvider(selectedProvider, pCfg, model, cfg.SystemPrompt, cfg.CommitPromptTemplate)
	if p == nil {
		fmt.Println(styleError.Render("Failed to init provider."))
		return
	}

	m := initialAISpinner(p, diff)
	pProgram := tea.NewProgram(m)
	finalModel, err := pProgram.Run()
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Error: %v", err)))
		return
	}
	finalState := finalModel.(aiSpinnerModel)
	if finalState.err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("AI Error: %v", finalState.err)))
		return
	}

	generatedMsg := finalState.msgResult

	// Review Loop
	parts := strings.SplitN(generatedMsg, "\n", 2)
	title := strings.TrimSpace(parts[0])
	description := ""
	if len(parts) > 1 {
		description = strings.TrimSpace(parts[1])
	}

	for {
		// Summary View
		// Create styles for content with wrapping
		width := 70
		labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).MarginBottom(0)
		contentStyle := lipgloss.NewStyle().PaddingLeft(2).Width(width).MaxWidth(width)
		
		boxStyle := lipgloss.NewStyle().
			MarginTop(1).
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1).
			Width(width + 4) // content + padding + border

		fmt.Println(boxStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				labelStyle.Render("Title:"),
				contentStyle.Render(title),
				"", // Spacer
				labelStyle.Render("Description:"),
				contentStyle.Render(description),
			),
		))

		var action string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Action").
					Options(
						huh.NewOption("Commit", "commit").Selected(true),
						huh.NewOption("Edit", "edit"),
						huh.NewOption("Cancel", "cancel"),
					).
					Value(&action),
			),
		)

		if err := form.Run(); err != nil {
			return
		}

		if action == "cancel" {
			return
		}
		if action == "commit" {
			break
		}
		if action == "edit" {
			f := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Title("Title").Value(&title),
					huh.NewInput().Title("Description").Value(&description),
				),
			)
			f.Run()
		}
	}

	finalMsg := fmt.Sprintf("%s\n\n%s", title, description)
	if err := git.Commit(finalMsg); err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Commit failed: %v", err)))
		return
	}
	fmt.Println(styleSuccess.Render("Committed successfully."))
}

func handlePush() {
	err := runSpinner("Pushing changes...", func() error {
		return git.Push()
	})
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Push failed: %v", err)))
	} else {
		fmt.Println(styleSuccess.Render("Pushed successfully."))
	}
}

func handlePull() {
	err := runSpinner("Pulling changes...", func() error {
		return git.Pull()
	})
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Pull failed: %v", err)))
	} else {
		fmt.Println(styleSuccess.Render("Pulled successfully."))
	}
}

func handleSync() {
	handleCommit()
	var confirm bool
	huh.NewForm(huh.NewGroup(huh.NewConfirm().Title("Push changes?").Value(&confirm))).Run()
	if confirm {
		handlePush()
	}
}

func handleInit() {
	err := runSpinner("Initializing AI-Git...", func() error {
		time.Sleep(500 * time.Millisecond) // UX pause
		root, err := git.GetRepoRoot()
		if err != nil {
			return fmt.Errorf("not a git repository")
		}
		repoConfigPath := filepath.Join(root, ".ai-git.yaml")
		if _, err := os.Stat(repoConfigPath); err == nil {
			return nil // Already exists
		}
		cfg, _ := config.LoadConfig()
		defaultProvider := "openai"
		if cfg != nil && cfg.DefaultProvider != "" {
			defaultProvider = cfg.DefaultProvider
		}
		content := fmt.Sprintf("enabled_provider: %s\ncommit_style: conventional\nlanguage: english\n", defaultProvider)
		return os.WriteFile(repoConfigPath, []byte(content), 0644)
	})

	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Init failed: %v", err)))
	} else {
		fmt.Println(styleSuccess.Render("Repository initialized."))
	}
}

func handleDoctor() {
	fmt.Println(styleTitle.Render("AI-Git Doctor"))

	check := func(label string, success bool, msg string) {
		icon := styleSuccess.Render("✓")
		if !success {
			icon = styleError.Render("✗")
		}
		fmt.Printf(" %s %s: %s\n", icon, label, msg)
	}

	// Git
	if git.IsRepo() {
		check("Git Repo", true, "Found")
	} else {
		check("Git Repo", false, "Not found")
	}

	// Config
	cfg, err := config.LoadConfig()
	if err != nil {
		check("Config", false, err.Error())
	} else {
		check("Config", true, "Loaded")
		if cfg.DefaultProvider == "" {
			check("Provider", false, "No default set")
		} else {
			check("Provider", true, cfg.DefaultProvider)
			pCfg, ok := cfg.Providers[cfg.DefaultProvider]
			if !ok {
				check("Setup", false, "Provider config missing")
			} else if pCfg.APIKey == "" && cfg.DefaultProvider != "ollama" {
				check("Auth", false, "API Key missing")
			} else {
				check("Auth", true, "API Key set")
			}
		}
	}
}

func handleConfig() {
	// If CLI args present, legacy mode
	if len(os.Args) > 2 {
		legacyConfig()
		return
	}

	// Interactive Mode
	fmt.Println(styleTitle.Render("Configuration"))

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println(styleError.Render("Failed to load config."))
		return
	}

	var provider string
	var apiKey string
	var model string

	// 1. Select Provider
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Default Provider").
				Options(
					huh.NewOption("OpenAI", "openai"),
					huh.NewOption("Gemini", "gemini"),
					huh.NewOption("Anthropic", "anthropic"),
					huh.NewOption("Ollama", "ollama"),
				).
				Value(&provider),
		),
	)

	if err := form.Run(); err != nil {
		return
	}

	// Load existing values
	pCfg := cfg.Providers[provider]
	apiKey = pCfg.APIKey
	model = pCfg.DefaultModel

	// 2. Configure Details
	// Use different fields depending on provider
	inputs := []huh.Field{
		huh.NewInput().
			Title("Default Model").
			Value(&model),
	}

	if provider != "ollama" {
		inputs = append([]huh.Field{
			huh.NewInput().
				Title("API Key").
				Value(&apiKey).
				Password(true),
		}, inputs...)
	}

	formDetails := huh.NewForm(
		huh.NewGroup(inputs...),
	)

	if err := formDetails.Run(); err != nil {
		return
	}

	// Save
	cfg.DefaultProvider = provider
	pCfg.APIKey = apiKey
	pCfg.DefaultModel = model
	if cfg.Providers == nil {
		cfg.Providers = make(map[string]config.ProviderConfig)
	}
	cfg.Providers[provider] = pCfg

	if err := cfg.Save(); err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Error saving: %v", err)))
	} else {
		fmt.Println(styleSuccess.Render("Configuration saved successfully."))
	}
}

func legacyConfig() {
	cfg, _ := config.LoadConfig()
	subCommand := os.Args[2]
	switch subCommand {
	case "set-provider":
		if len(os.Args) < 4 {
			return
		}
		cfg.DefaultProvider = os.Args[3]
	case "set-key":
		if len(os.Args) < 5 {
			return
		}
		pName := os.Args[3]
		pCfg := cfg.Providers[pName]
		pCfg.APIKey = os.Args[4]
		cfg.Providers[pName] = pCfg
	case "set-model":
		if len(os.Args) < 5 {
			return
		}
		pName := os.Args[3]
		pCfg := cfg.Providers[pName]
		pCfg.DefaultModel = os.Args[4]
		cfg.Providers[pName] = pCfg
	}
	cfg.Save()
}

func parseGitStatusFiles(status string) []string {
	var files []string
	lines := strings.Split(status, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if len(trimmed) > 3 {
			filePart := strings.TrimSpace(line[2:])
			files = append(files, filePart)
		}
	}
	return files
}
