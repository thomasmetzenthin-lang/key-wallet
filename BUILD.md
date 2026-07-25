# key-wallet Build & Architecture Guide

This document details the compilation pipeline, build tag configurations, and version management strategy for key-wallet.

---

## 1. Build Tag Variants

key-wallet uses Go build constraints (//go:build) to separate core proxy execution from desktop GUI/OS hooks.

### A. Headless / Server Daemon (Default)
Minimal memory footprint, CLI-only mode. Ideal for background services, Docker containers, or remote servers.

\\\powershell
go build -o key-wallet.exe ./cmd/key-wallet
\\\

### B. Desktop Variant
Enables desktop-specific integrations (clipboard auto-clearing, native notifications, system tray hooks).

\\\powershell
go build -tags desktop -o key-wallet-desktop.exe ./cmd/key-wallet
\\\

---

## 2. Iteration & Versioning Strategy

- **Git Tags:** Official snapshot releases are tagged with semver (v0.1.0, v0.2.0).
- **Internal Development Vault:** Private architectural notes, chat logs, and binary backups are stored locally in .dev_archive/ (git-ignored).
