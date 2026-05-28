package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/eliau2005/ai-git/internal/config"
	"github.com/eliau2005/ai-git/internal/git"
	"github.com/eliau2005/ai-git/internal/provider"
	"github.com/eliau2005/ai-git/internal/rag"
)

func handleIndex() {
	fmt.Println(styleTitle.Render("Index Repository for AI Chat"))

	root, err := git.GetRepoRoot()
	if err != nil {
		fmt.Println(styleError.Render("Not in a git repository."))
		return
	}

	p, ok := getActiveProvider().(provider.Embedder)
	if !ok {
		fmt.Println(styleError.Render("Current provider does not support embeddings."))
		return
	}

	store, _ := rag.LoadStore(root)
	store.Chunks = nil // Clear existing chunks for re-index

	var count int
	rules, _ := git.LoadIgnoreRules(root)

	fmt.Println(styleSubtle.Render("Scanning and embedding files... This may take a moment due to API rate limits."))

	err = filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			// Skip .git directory
			if info != nil && info.IsDir() && info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, _ := filepath.Rel(root, path)
		if git.ShouldIgnore(relPath, rules) {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".go" && ext != ".md" && ext != ".txt" && ext != ".js" && ext != ".ts" {
			return nil // Skip non-code/docs
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// Simple chunking
		text := string(content)
		if len(text) > 4000 {
			text = text[:4000] // simple truncation
		}

		fmt.Printf("Indexing %s... ", relPath)
		emb, err := p.GenerateEmbedding(text)
		if err == nil && len(emb) > 0 {
			store.AddChunk(rag.Chunk{
				ID:        relPath,
				FilePath:  relPath,
				Content:   text,
				Embedding: emb,
			})
			count++
			fmt.Println("Done")
			time.Sleep(2 * time.Second) // Rate limit prevention
		} else {
			fmt.Println("Failed:", err)
		}
		return nil
	})

	if err != nil {
		fmt.Println(styleError.Render(fmt.Sprintf("Error during indexing: %v", err)))
		return
	}

	store.Save(root)
	fmt.Println(styleSuccess.Render(fmt.Sprintf("Successfully indexed %d files.", count)))
}

func handleChat() {
	fmt.Println(styleTitle.Render("Chat with your Repository"))

	root, err := git.GetRepoRoot()
	if err != nil {
		fmt.Println(styleError.Render("Not in a git repository."))
		return
	}

	activeProv := getActiveProvider()
	chatter, okChat := activeProv.(provider.Chatter)
	embedder, okEmbed := activeProv.(provider.Embedder)

	if !okChat || !okEmbed {
		fmt.Println(styleError.Render("Current provider does not fully support Chat and Embeddings."))
		return
	}

	store, err := rag.LoadStore(root)
	if err != nil || len(store.Chunks) == 0 {
		fmt.Println(styleError.Render("No index found. Please run 'ai-git index' first."))
		return
	}

	scanner := bufio.NewScanner(os.Stdin)
	userStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	aiStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

	for {
		fmt.Print(userStyle.Render("\nYou: "))
		if !scanner.Scan() {
			break
		}
		query := strings.TrimSpace(scanner.Text())
		if query == "exit" || query == "quit" {
			break
		}
		if query == "" {
			continue
		}

		queryEmb, err := embedder.GenerateEmbedding(query)
		if err != nil {
			fmt.Println(styleError.Render(fmt.Sprintf("Failed to embed query: %v", err)))
			continue
		}

		results := store.Search(queryEmb, 3)
		var contextBuilder strings.Builder
		for _, r := range results {
			contextBuilder.WriteString(fmt.Sprintf("--- File: %s ---\n%s\n", r.Chunk.FilePath, r.Chunk.Content))
		}

		fmt.Print(aiStyle.Render("\nAI: "))
		err = chatter.AskChatStream(query, contextBuilder.String(), func(chunk string) {
			fmt.Print(chunk)
		})
		fmt.Println()

		if err != nil {
			fmt.Println(styleError.Render(fmt.Sprintf("\nError: %v", err)))
		}
	}
}

func getActiveProvider() provider.Provider {
	cfg, _ := config.LoadConfig()
	root, _ := git.GetRepoRoot()
	repoCfg, _ := config.LoadRepoConfig(root)

	selectedProvider := cfg.DefaultProvider
	if repoCfg != nil && repoCfg.EnabledProvider != "" {
		selectedProvider = repoCfg.EnabledProvider
	}
	pCfg := cfg.Providers[selectedProvider]

	model := pCfg.DefaultModel
	if repoCfg != nil && repoCfg.ModelOverride != "" {
		model = repoCfg.ModelOverride
	}

	factory := &provider.ProviderFactory{}
	return factory.GetProvider(selectedProvider, pCfg, model, cfg.SystemPrompt, cfg.CommitPromptTemplate)
}
