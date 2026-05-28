package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/eliau2005/ai-git/internal/git"
	"github.com/eliau2005/ai-git/internal/provider"
)

func handleRefactor(args []string) {
	fmt.Println(styleTitle.Render("Autonomous Agent: Code Refactor"))

	if len(args) < 1 {
		fmt.Println(styleError.Render("Usage: ai-git refactor <file> [prompt]"))
		return
	}

	targetFile := args[0]
	var prompt string
	if len(args) > 1 {
		prompt = strings.Join(args[1:], " ")
	} else {
		// Prompt the user interactively
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("What should I refactor or change?").
					Placeholder("e.g. Convert to Typescript, Optimize the loop, Add comments...").
					Value(&prompt),
			),
		)
		if err := form.Run(); err != nil {
			return
		}
	}

	if prompt == "" {
		fmt.Println(styleError.Render("Prompt cannot be empty."))
		return
	}

	// Read file
	contentBytes, err := os.ReadFile(targetFile)
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Could not read %s: %v", targetFile, err)))
		return
	}

	activeProv := getActiveProvider()
	refactorer, ok := activeProv.(provider.CodeRefactorer)
	if !ok {
		fmt.Println(styleError.Render("Current provider does not support autonomous refactoring."))
		return
	}

	var newCode string
	err = runSpinner(fmt.Sprintf("Agent is refactoring %s...", targetFile), func() error {
		res, e := refactorer.RefactorCode(prompt, string(contentBytes))
		newCode = res
		return e
	})

	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Failed to refactor code: %v", err)))
		return
	}

	var action string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("AI completed the refactor. What next?").
				Options(
					huh.NewOption("Review changes (git diff)", "review"),
					huh.NewOption("Accept and Stage", "accept"),
					huh.NewOption("Discard", "discard"),
				).
				Value(&action),
		),
	)

	if err := form.Run(); err != nil {
		return
	}

	if action == "discard" {
		fmt.Println(styleSubtle.Render("Refactor discarded."))
		return
	}

	// Write to file temporarily to diff or stage
	err = os.WriteFile(targetFile, []byte(newCode), 0644)
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Failed to write to %s: %v", targetFile, err)))
		return
	}

	if action == "review" {
		// We can show a git diff of the file
		diff, _ := git.Diff(targetFile)
		fmt.Println("\n" + diff + "\n")

		var confirm bool
		formConfirm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Keep and stage these changes?").
					Value(&confirm),
			),
		)
		if formConfirm.Run() == nil && confirm {
			git.Add(targetFile)
			fmt.Println(styleSuccess.Render("Changes staged successfully! 🚀"))
		} else {
			// Revert file
			os.WriteFile(targetFile, contentBytes, 0644)
			fmt.Println(styleSubtle.Render("Reverted changes."))
		}
	} else if action == "accept" {
		git.Add(targetFile)
		fmt.Println(styleSuccess.Render(fmt.Sprintf("Successfully refactored and staged %s! 🚀", targetFile)))
	}
}
