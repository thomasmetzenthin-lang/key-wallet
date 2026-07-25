# key-wallet ???

key-wallet is a lightweight, zero-disk local proxy daemon written in Go. It securely holds API keys in memory and routes local LLM client traffic (OpenRouter, Anthropic, OpenAI, Gemini) through 127.0.0.1:9090 without writing credentials to disk or local configuration files.

---

## Section 1: Quickstart (Operational Baseline)

### 1. Build the Binary
`powershell
go build -o key-wallet.exe ./cmd/key-wallet
.\key-wallet.exe


## Troubleshooting & Integration Notes

- **Container Host Resolution:** When integrating containerized frontends (e.g., Dockerized LibreChat), configure the `baseURL` to target `http://host.docker.internal:9090` instead of `localhost` or `127.0.0.1`.
- **Dynamic Model Fetching (`fetch: true`):** Setting `fetch: true` causes client applications to poll `http://host.docker.internal:9090/proxy/openrouter/v1/models` on startup. If the daemon is inactive or non-responsive, frontends may suppress the custom endpoint from the UI selection menu entirely. Ensure the daemon is running before starting client stacks.
