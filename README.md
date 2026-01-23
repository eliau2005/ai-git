# AI-Git CLI

![License: MIT](https://img.shields.io/badge/License-MIT-purple.svg)
![Platform: Linux](https://img.shields.io/badge/Platform-Linux-blue.svg)
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

## 🚀 Installation Guide (Ubuntu / Linux)

### Prerequisites

Ensure you have the following installed on your system:

1.  **Git**: Version control system.
    ```bash
    sudo apt update
    sudo apt install git
    ```
2.  **Go (Golang)**: Version 1.21 or higher is required to build from source.
    ```bash
    # Remove existing Go installation
    sudo rm -rf /usr/local/go
    
    # Download and extract Go (check golang.org for the latest version)
    wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
    
    # Add Go to your PATH (add this to ~/.bashrc or ~/.zshrc)
    export PATH=$PATH:/usr/local/go/bin
    source ~/.bashrc
    ```

### Build & Install

1.  **Clone the Repository**
    ```bash
    git clone https://github.com/eliau2005/ai-git.git
    cd ai-git
    ```

2.  **Download Dependencies**
    The project uses Go modules. This command will fetch all required libraries (CharmBracelet, etc.).
    ```bash
    go mod tidy
    ```

3.  **Build the Binary**
    Compile the source code into an executable.
    ```bash
    go build -o ai-git cmd/ai-git/main.go
    ```

4.  **Install Globally**
    Move the binary to your system's bin directory to access it from anywhere.
    ```bash
    sudo mv ai-git /usr/local/bin/
    ```

5.  **Verify Installation**
    ```bash
    ai-git version
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