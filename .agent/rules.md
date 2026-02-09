---
scope: project
extends: /home/antigravity/.antigravity/agent_process_rules.md
---

# Project Rules (v2026.02)

## 1. SECURITY MANDATES (Non-Negotiable)
> [!CAUTION]
> These rules override any conflicting instructions.

### 1.1 GPG Black Box
- **STOP AND REPORT** if GPG signing fails.
- **DO NOT** attempt to generate, export, or modify GPG keys.
- **DO NOT** touch `~/.gnupg` directory.

### 1.2 Credential Hygiene
- **NEVER** commit secrets (check `.gitignore`).
- **REDACT** all tokens/keys from command output and logs.
- **USE** existing auth mechanisms; do not invent new ones.

## 2. REMOTE REPOSITORY STANDARDS
- **GIT COMMIT -S**: Required for all commits if `git remote` exists.
- **PUSH VERIFICATION**: Ensure commits are verified on GitHub.

## 3. PROJECT EXTENSIONS
[Project specific rules will be appended here or in project_rules.md]
