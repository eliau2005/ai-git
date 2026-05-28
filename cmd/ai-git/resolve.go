package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/eliau2005/ai-git/internal/git"
	"github.com/eliau2005/ai-git/internal/provider"
)

func handleResolve() {
	fmt.Println(styleTitle.Render("AI Merge Conflict Resolver"))

	conflicts, err := git.GetConflictingFiles()
	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Failed to get conflicts: %v", err)))
		return
	}

	if len(conflicts) == 0 {
		fmt.Println(styleSuccess.Render("No merge conflicts found! 🎉"))
		return
	}

	activeProv := getActiveProvider()
	resolver, ok := activeProv.(provider.ConflictResolver)
	if !ok {
		fmt.Println(styleError.Render("Current provider does not support conflict resolution."))
		return
	}

	for _, file := range conflicts {
		if file == "" {
			continue
		}

		fmt.Println(styleSubtle.Render(fmt.Sprintf("\nAnalyzing conflict in: %s", file)))

		contentBytes, err := os.ReadFile(file)
		if err != nil {
			fmt.Println(styleError.Render(fmt.Sprintf("Could not read file %s: %v", file, err)))
			continue
		}
		content := string(contentBytes)

		var resolvedContent string
		err = runSpinner("AI is resolving conflicts...", func() error {
			res, e := resolver.ResolveConflict(content)
			resolvedContent = res
			return e
		})

		if err != nil {
			fmt.Println(styleError.Render(fmt.Sprintf("Failed to resolve %s: %v", file, err)))
			continue
		}

		var action string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title(fmt.Sprintf("AI proposed a resolution for %s", file)).
					Options(
						huh.NewOption("Accept and Save", "accept"),
						huh.NewOption("Discard", "discard"),
					).
					Value(&action),
			),
		)

		if err := form.Run(); err != nil {
			continue
		}

		if action == "accept" {
			err = os.WriteFile(file, []byte(resolvedContent), 0644)
			if err != nil {
				fmt.Println(styleError.Render(fmt.Sprintf("Failed to save resolved file %s: %v", file, err)))
			} else {
				// automatically stage the resolved file
				git.Add(file)
				fmt.Println(styleSuccess.Render(fmt.Sprintf("Successfully resolved and staged %s!", file)))
			}
		} else {
			fmt.Println(styleSubtle.Render("Skipped resolving " + file))
		}
	}
}
