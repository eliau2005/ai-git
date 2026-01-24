# AI-Git CLI 🤖

AI-Git is an advanced, interactive CLI tool designed to supercharge your Git workflow.
It automatically writes smart, organized, and conventional commit messages by analyzing your code changes using Artificial Intelligence.

**Supports:** OpenAI (GPT-4), Google Gemini, Claude, and Ollama (Local).

## ✨ Key Features

- ⚡ **Fast & Lightweight**: Operates directly from your terminal.
- 🎨 **Modern Interface (TUI)**: Visual file selection, loading animations, and colorful menus.
- 🧠 **Intelligent**: Analyzes git diff to understand exactly what you changed.
- ✏️ **Full Control**: Edit the AI-generated message before committing.
- ⚙️ **Easy Configuration**: Interactive setup wizard (`ai-git config`).

## 🚀 Installation Instructions

### Prerequisites
- Git installed.
- Go (version 1.21 or higher).

### 🌟 Option 1: Quick Install (Recommended)
If you have Go configured on your machine, this is the fastest method (Works on Windows & Linux):
```bash
go install github.com/eliau2005/ai-git/cmd/ai-git@latest
```
*Note: Ensure your Go Bin directory is in your PATH.*

### 🛠️ Option 2: Manual Install (Build from Source)
Use this method if you cloned the repository (`git clone`).

#### 🐧 Linux / Mac
1. **Build the binary:**
   ```bash
   go build -o ai-git cmd/ai-git/main.go
   ```
2. **Move it to a global directory:**
   ```bash
   sudo mv ai-git /usr/local/bin/
   ```
3. **Verify:**
   ```bash
   ai-git version
   ```

#### 🪟 Windows (PowerShell)
1. **Build the binary:**
   ```powershell
   go build -o ai-git.exe cmd/ai-git/main.go
   ```
2. **Create a tools folder and move the file (Recommended):**
   ```powershell
   # Create folder
   New-Item -ItemType Directory -Force -Path "C:\Tools"

   # Move file
   Move-Item -Path .\ai-git.exe -Destination "C:\Tools\ai-git.exe" -Force
   ```
3. **Add to PATH (so you can run it from anywhere):**
   Run this command only once:
   ```powershell
   [Environment]::SetEnvironmentVariable("PATH", $env:PATH + ";C:\Tools", "User")
   ```
   *Now close your terminal and reopen it.*

## 🔄 How to Update?
New version released? Here is how to update in seconds:

- **If installed via `go install`:**
  Simply run the command again:
  ```bash
  go install github.com/eliau2005/ai-git/cmd/ai-git@latest
  ```

- **If installed manually:**
  Navigate to the project folder and run:
  
  **Linux:**
  ```bash
  git pull
  go build -o ai-git cmd/ai-git/main.go
  sudo mv ai-git /usr/local/bin/
  ```
  
  **Windows:**
  ```powershell
  git pull
  go build -o ai-git.exe cmd/ai-git/main.go
  Move-Item -Path .\ai-git.exe -Destination "C:\Tools\ai-git.exe" -Force
  ```

## 🎮 Usage Guide

### 1. Initial Setup
Before the first use, configure your AI provider (API Key):
```bash
ai-git config
```
An interactive menu will open to select a provider (e.g., Gemini or OpenAI) and enter your key.

### 2. Initialize Project
Inside your project folder:
```bash
ai-git init
```

### 3. Daily Workflow
- **Stage Files:**
  ```bash
  ai-git add
  ```
  Opens a selection menu (Checklist) of changed files.
- **Create Commit:**
  ```bash
  ai-git commit
  ```
  AI analyzes changes, suggests a message, and lets you edit it.
- **All in One (Add + Commit + Push):**
  ```bash
  ai-git sync
  ```

### 4. Troubleshooting
Something not working? Run the doctor:
```bash
ai-git doctor
```

## 📄 License
This project is released under the MIT License.

Created by **eliau2005**.
