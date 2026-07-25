# key-wallet ???

key-wallet is a lightweight, zero-disk local proxy daemon written in Go. It securely holds API keys in memory and routes local LLM client traffic (OpenRouter, Anthropic, OpenAI, Gemini) through 127.0.0.1:9090 without writing credentials to disk or local configuration files.

---

## Section 1: Quickstart (Operational Baseline)

### 1. Build the Binary
`powershell
go build -o key-wallet.exe ./cmd/key-wallet
.\key-wallet.exe

