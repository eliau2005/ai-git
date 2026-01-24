# AI-Git CLI

![License: MIT](https://img.shields.io/badge/License-MIT-purple.svg)
![Platform: Linux & Windows](https://img.shields.io/badge/Platform-Linux%20%7C%20Windows-blue.svg)
![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8.svg)

**AI-Git** is a modern, interactive CLI tool that supercharges your Git workflow with Artificial Intelligence. It automates the tedious parts of version control—like writing commit messages—by analyzing your changes and generating high-quality, conventional commit messages using your preferred AI provider (OpenAI, Google Gemini, Anthropic Claude, or Ollama).

Built with Go and the [Charm](https://charm.sh/) TUI libraries, AI-Git offers a beautiful, terminal-based user interface for managing your repository.

## ✨ Features

*   **🤖 Multi-Provider Support:** Works with OpenAI (GPT-4), Google Gemini, Anthropic Claude, and local models via Ollama.
*   **🖥️ Interactive TUI:** rich, modern terminal interface for all commands.
*   **📝 Smart Commit Generation:** Automatically analyzes `git diff` to generate concise, conventional commit messages.
*   **✅ Interactive Staging:** Selectively stage files with a checklist interface.
*   **✏️ Review & Edit:** Review, edit (Title/Description), and confirm AI-generated messages before committing.
*   **⚙️ Easy Configuration:** Guided, interactive configuration menu—no manual YAML editing required.
*   **🏥 System Doctor:** Built-in diagnostic tool to verify your environment setup.

## 🚀 Installation Guide

### Prerequisites

Ensure you have the following installed on your system:
1.  **Git**: Version control system.
2.  **Go (Golang)**: Version 1.21 or higher is required to build from source.

### Ubuntu / Linux

1.  **Clone the Repository**
    ```bash
    git clone https://github.com/eliau2005/ai-git.git
    cd ai-git
    ```

2.  **Download Dependencies**
    ```bash
    go mod tidy
    ```

3.  **Build the Binary**
    ```bash
    go build -o ai-git cmd/ai-git/main.go
    ```

4.  **Install Globally**
    ```bash
    sudo mv ai-git /usr/local/bin/
    ```

5.  **Verify Installation**
    ```bash
    ai-git version
    ```

### Windows

1.  **Clone the Repository**
    Open PowerShell (Run as Administrator for path setup):
    ```powershell
    git clone https://github.com/eliau2005/ai-git.git
    cd ai-git
    ```

2.  **Download Dependencies**
    ```powershell
    go mod tidy
    ```

3.  **Build the Binary**
    ```powershell
    go build -o ai-git.exe cmd/ai-git/main.go
    ```

4.  **Install Globally**
    Create a folder for your tools and add it to your User PATH:
    ```powershell
    # Create the folder
    New-Item -ItemType Directory -Force -Path "C:\Tools"
    
    # Move the binary
    Move-Item -Path .\ai-git.exe -Destination "C:\Tools\ai-git.exe" -Force

    # Add to PATH (Permanent for current user)
    $OldPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($OldPath -notlike "*C:\Tools*") {
        [Environment]::SetEnvironmentVariable("PATH", "$OldPath;C:\Tools", "User")
        Write-Host "PATH updated. Please restart your terminal." -ForegroundColor Green
    }
    ```

5.  **Verify Installation**
    Restart PowerShell and run:
    ```powershell
    ai-git version
    ```

## 🔄 Updating to a Newer Version

To update AI-Git to the latest version, follow these steps:

1.  **Navigate to the source directory:**
    ```bash
    cd /path/to/ai-git
    ```

2.  **Pull the latest changes:**
    ```bash
    git pull origin main
    ```

3.  **Rebuild and Reinstall:**

    **Linux/Ubuntu:**
    ```bash
    go build -o ai-git cmd/ai-git/main.go
    sudo mv ai-git /usr/local/bin/
    ```

    **Windows:**
    ```powershell
    go build -o ai-git.exe cmd/ai-git/main.go
    Move-Item -Path .\ai-git.exe -Destination "C:\Tools\ai-git.exe" -Force
    ```

## 🛠️ Usage

### 1. Configuration
Before using the tool, set up your AI provider.
```bash
ai-git config
```
This launches an interactive menu where you can:
*   Select your provider (OpenAI, Gemini, Anthropic, Ollama).
*   Enter your API Key.
*   Set your preferred model (e.g., `gpt-4o`, `gemini-1.5-flash`).

### 2. Initialize Repository
In your project root, run:
```bash
ai-git init
```
This creates a local `.ai-git.yaml` configuration file for the repository.

### 3. Workflow
The typical workflow replaces standard git commands with their AI-powered counterparts:

*   **Stage Files:**
    ```bash
    ai-git add
    # Opens an interactive checklist to select files
    ```

*   **Generate Commit:**
    ```bash
    ai-git commit
    ```
    *   Scans staged changes (or asks to stage if empty).
    *   Generates a commit message using AI.
    *   Presents a review screen to Edit (Title/Description), Commit, or Cancel.

*   **Sync (Commit & Push):**
    ```bash
    ai-git sync
    # Performs Add -> Commit -> Push in one flow
    ```

### 4. Diagnostics
If something isn't working, run the doctor command to check your setup.
```bash
ai-git doctor
```

## 📦 Dependencies

This project relies on the following open-source Go libraries:

*   **[Bubble Tea](https://github.com/charmbracelet/bubbletea):** The fun, functional, and stateful terminal apps framework.
*   **[Lip Gloss](https://github.com/charmbracelet/lipgloss):** Style definitions for nice terminal layouts.
*   **[Huh](https://github.com/charmbracelet/huh):** A simple, powerful library for forms and fields in the terminal.
*   **[Bubbles](https://github.com/charmbracelet/bubbles):** TUI components (spinners, etc.).
*   **[Go YAML](https://github.com/go-yaml/yaml):** YAML support for the Go language.

## 🤝 Contributing

Contributions are welcome! Please follow these steps:
1.  Fork the repository.
2.  Create a feature branch (`git checkout -b feature/AmazingFeature`).
3.  Commit your changes.
4.  Push to the branch.
5.  Open a Pull Request.

## 📄 License

This project is open-source and available under the **MIT License**.

Created by **[eliau2005](https://github.com/eliau2005)**.