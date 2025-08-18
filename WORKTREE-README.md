# Worktree Setup Guide

This document explains how to configure git settings for worktrees to ensure proper attribution across multiple working trees.

## Prerequisites

First, enable the worktree config extension to allow per-worktree configurations:

```bash
# Run this once from any worktree (enables worktree-specific configs)
git config extensions.worktreeConfig true
```

## Git Configuration for Worktrees

### Main Repository Configuration

Set the configuration for your main repository:

```bash
# In the main repository directory
cd /path/to/main/repo
git config --local user.name "Your Name"
git config --local user.email "your.email@example.com"
```

### Worktree-Specific Configuration

For each worktree, set unique user credentials:

```bash
# In each worktree directory, use --worktree flag
cd /path/to/worktree

# For Claude/AI work:
git config --worktree user.name "Claude"
git config --worktree user.email "noreply@anthropic.com"

# For Gemini work:
git config --worktree user.name "Gemini"
git config --worktree user.email "gemini@google.com"
```

### Verify Configuration

```bash
# Check current settings (from any worktree or main repo)
git config user.name
git config user.email

# View worktree-specific config only
git config --worktree --list

# View local repository config only
git config --local --list

# See where config values are coming from
git config --show-origin user.name
git config --show-origin user.email
```

## Understanding the Configuration Hierarchy

Git uses this precedence order:

1. `--worktree` (worktree-specific, stored in `.git/worktrees/<name>/config.worktree`)
2. `--local` (repository-specific, stored in `.git/config`)
3. `--global` (user-specific, stored in `~/.gitconfig`)
4. `--system` (system-wide)

## Notes

- **Main repository**: Uses `--local` config settings
- **Worktrees**: Use `--worktree` config settings that override the main repo's local config
- Each worktree maintains its own identity without affecting others
- The `extensions.worktreeConfig` setting only needs to be enabled once per repository
- These settings persist until manually changed or the worktree is deleted
- All commits made in each location will use their respective credentials

## Environment Configuration

For parallel development, create environment-specific files in each worktree to avoid conflicts:

```bash
# Example: Create .env.local with different ports
# Main repo: PORT=3000
# Claude worktree: PORT=3001
# Gemini worktree: PORT=3002
```
