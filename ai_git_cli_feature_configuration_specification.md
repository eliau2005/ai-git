# AI-Git CLI – Specification

## 1. Overview

AI-Git is a cross-platform CLI tool (Windows & Linux) that integrates Git workflows with multiple AI providers to automate commit message generation and streamline common Git operations.

The tool runs in the current working directory and operates entirely through terminal commands.

---

## 2. Core Objectives

* Provide a unified CLI interface for Git + AI
* Support multiple AI providers (cloud & local)
* Generate high-quality commit messages automatically
* Be fast, portable, and dependency-free (single binary)
* Be extensible (providers, models, commands)

---

## 3. Supported Platforms

* Linux (x64)
* Windows (10/11)
* Mac OS

---

## 4. Technology Stack

* Language: Go (Golang)
* Distribution: Single compiled binary
* Target OS: Linux, Windows, Mac OS
* Configuration format: YAML (primary), JSON (optional)

---

## 5. AI Providers

### 5.1 Supported Providers (Built-in)

* OpenAI (ChatGPT)
* Google Gemini
* Anthropic Claude
* Ollama (Local models)

---

### 5.2 Provider Capabilities

| Capability        | Cloud Providers | Ollama |
| ----------------- | --------------- | ------ |
| API Key Required  | Yes             | No     |
| Internet Required | Yes             | No     |
| Custom Models     | Yes             | Yes    |
| Default Models    | Yes             | Yes    |

---

### 5.3 Default Models (Initial)

#### OpenAI

* gpt-4.1
* gpt-4o
* gpt-4o-mini

#### Gemini

* gemini-1.5-pro
* gemini-1.5-flash

#### Claude

* claude-3-opus
* claude-3-sonnet

#### Ollama

* llama3
* mistral

---

### 5.4 Custom Models

Users can:

* Add additional model names per provider
* Set a default model per provider

---

## 6. Configuration Management

### 6.1 Global Configuration

Location:

* Linux: `~/.config/ai-git/config.yaml`

Contains:

* Default AI provider
* Default model per provider
* API keys (encrypted or obfuscated)
* Output preferences

---

### 6.2 Repository Configuration

Location:

* `.ai-git.yaml` (root of repository)

Contains:

* Enabled provider for the repo
* Model override
* Commit message style
* Language preference

---

## 7. Git Features

### 7.1 Repository Management

* Initialize repository as AI-Git enabled
* Validate Git repository existence

---

### 7.2 Git Commands

Supported operations:

* `status` – show repository status
* `pull` – fetch and merge remote changes
* `add` – stage changes
* `commit` – create commit with AI-generated message
* `push` – push commits to remote

---

### 7.3 Combined Workflow Command

Single command executing:

1. git status
2. git add (interactive or all)
3. AI-generated commit message
4. git commit
5. git push

---

## 8. AI Commit Message Generation

### 8.1 Input Context

AI receives:

* `git diff --staged` or `git diff`
* File names changed
* Change statistics

---

### 8.2 Output Rules

* Clear, concise commit message
* Conventional commit style (optional)
* Configurable language (default: English)

---

### 8.3 Prompt Customization

Users can configure:

* Commit style (short / detailed)
* Prefixes (feat, fix, refactor, etc.)
* Max length

---

## 9. CLI Commands

### 9.1 Global Commands

* `ai-git version`
* `ai-git config`
* `ai-git doctor` (validate setup)

---

### 9.2 Repository Commands

* `ai-git init`
* `ai-git status`
* `ai-git pull`
* `ai-git add`
* `ai-git commit`
* `ai-git push`
* `ai-git sync` (status → add → commit → push)

---

## 10. Error Handling & Validation

* Missing API key detection
* Unsupported model validation
* Git repository validation
* Ollama availability check

---

## 11. UX & CLI Behavior

* Clear terminal output
* Colorized status messages
* Dry-run mode
* Verbose / debug mode

---

## 12. Security Considerations

* Do not log API keys
* Config file permissions validation
* Optional environment variable support for secrets

---

## 13. Extensibility

* Provider interface abstraction
* Easy addition of new AI providers
* Plugin-ready command structure

---

