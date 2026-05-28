package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/eliau2005/ai-git/internal/provider"
)

func handleFix(args []string) {
	fmt.Println(styleTitle.Render("Smart Diagnostics & Auto Fix"))

	var errorText string

	// Check if data is piped
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		bytes, _ := io.ReadAll(os.Stdin)
		errorText = string(bytes)
	}

	// Also check arguments
	if len(args) > 0 {
		errorText += "\n" + strings.Join(args, " ")
	}

	errorText = strings.TrimSpace(errorText)

	if errorText == "" {
		fmt.Println(styleError.Render("Usage: pipe a failed command output to this tool (e.g., `go build 2>&1 | ai-git fix`) or pass the error as an argument."))
		return
	}

	activeProv := getActiveProvider()
	chatter, ok := activeProv.(provider.Chatter)
	if !ok {
		fmt.Println(styleError.Render("Current provider does not support chat/diagnostics."))
		return
	}

	prompt := "I encountered the following error while working in my repository. Please analyze the error, explain what might be causing it, and provide the exact fix or shell commands to resolve it:\n\n" + errorText

	fmt.Println(styleSubtle.Render("\nAnalyzing error...\n"))

	err := chatter.AskChatStream(prompt, "", func(chunk string) {
		fmt.Print(chunk)
	})
	fmt.Println()

	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("\nDiagnostics failed: %v", err)))
	}
}
