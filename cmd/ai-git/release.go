package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/eliau2005/ai-git/internal/provider"
)

func handleRelease() {
	fmt.Println(styleTitle.Render("Semantic Release & Changelog Generator"))

	// Find the last tag
	cmdTag := exec.Command("git", "describe", "--tags", "--abbrev=0")
	var outTag bytes.Buffer
	cmdTag.Stdout = &outTag
	err := cmdTag.Run()

	var commitRange string
	var lastTag string

	if err != nil {
		// No tag found, use all commits
		commitRange = "HEAD"
		fmt.Println(styleSubtle.Render("No previous tags found. Analyzing all commits..."))
	} else {
		lastTag = strings.TrimSpace(outTag.String())
		commitRange = fmt.Sprintf("%s..HEAD", lastTag)
		fmt.Println(styleSubtle.Render(fmt.Sprintf("Analyzing commits since %s...", lastTag)))
	}

	// Get commit messages
	cmdLog := exec.Command("git", "log", commitRange, "--pretty=format:%s")
	var outLog bytes.Buffer
	cmdLog.Stdout = &outLog
	err = cmdLog.Run()
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Failed to get commits: %v", err)))
		return
	}

	commits := strings.TrimSpace(outLog.String())
	if commits == "" {
		fmt.Println(styleSuccess.Render("No new commits to release!"))
		return
	}

	activeProv := getActiveProvider()
	chatter, ok := activeProv.(provider.Chatter)
	if !ok {
		fmt.Println(styleError.Render("Current provider does not support chat/changelog generation."))
		return
	}

	prompt := "You are a release manager. Group the following commit messages into a beautifully formatted Markdown Changelog. Categorize them into '✨ Features', '🐛 Bug Fixes', and '🛠️ Maintenance' (or similar). Do NOT include markdown codeblocks around your entire response. Here are the commits:\n\n" + commits

	var changelog string

	fmt.Println(styleSubtle.Render("Generating Changelog..."))
	fmt.Println(strings.Repeat("-", 40))
	
	var sb strings.Builder
	err = chatter.AskChatStream(prompt, "", func(chunk string) {
		fmt.Print(chunk)
		sb.WriteString(chunk)
	})
	fmt.Println("\n" + strings.Repeat("-", 40))

	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Error generating changelog: %v", err)))
		return
	}

	changelog = sb.String()

	var confirm bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Save this to CHANGELOG.md?").
				Value(&confirm),
		),
	)

	if err := form.Run(); err != nil || !confirm {
		fmt.Println(styleSubtle.Render("Skipped saving changelog."))
		return
	}

	// Append or Create CHANGELOG.md
	existing, _ := os.ReadFile("CHANGELOG.md")
	newContent := changelog + "\n\n" + string(existing)
	err = os.WriteFile("CHANGELOG.md", []byte(newContent), 0644)
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Failed to save CHANGELOG.md: %v", err)))
	} else {
		fmt.Println(styleSuccess.Render("Saved to CHANGELOG.md! 🚀"))
	}
}
