package main

import (
"bufio"
"fmt"
"log"
"net/http"
"os"
"strings"

"key-wallet/pkg/keyring"
"key-wallet/pkg/proxy"
)

func main() {
fmt.Println("==================================================")
fmt.Println("            key-wallet Security Daemon            ")
fmt.Println("==================================================")

// Execute build-variant feature initialization
initDesktopFeatures()

fmt.Println("Paste all keys below (separated by newlines or as a single block).")
fmt.Println("Type 'DONE' or press Enter on an empty line when finished:\n")

store, err := keyring.NewSecureStore()
if err != nil {
log.Fatalf("Failed to initialize secure store: %v", err)
}

scanner := bufio.NewScanner(os.Stdin)
keysLoaded := 0

for {
if !scanner.Scan() {
break
}
line := strings.TrimSpace(scanner.Text())

if line == "" || strings.ToUpper(line) == "DONE" {
if keysLoaded > 0 {
break
}
fmt.Println("[!] No keys provided. Please paste at least one key or press Ctrl+C to exit.")
continue
}

// Detect provider based on key prefix
var provider string
switch {
case strings.HasPrefix(line, "sk-or-v1-"):
provider = "openrouter"
case strings.HasPrefix(line, "sk-ant-"):
provider = "anthropic"
case strings.HasPrefix(line, "AIzaSy"):
provider = "gemini"
case strings.HasPrefix(line, "sk-"):
provider = "openai"
default:
fmt.Printf("[?] Unrecognized format for line starting with '%s...'. Skipping.\n", line[:min(len(line), 8)])
continue
}

// Store in encrypted RAM
if err := store.Set(provider, []byte(line)); err != nil {
fmt.Printf("[!] Error storing key for %s: %v\n", provider, err)
continue
}

fmt.Printf("[+] Loaded %s key into RAM.\n", provider)
keysLoaded++
}

// Clear terminal screen buffer to prevent visual leakage
fmt.Print("\033[H\033[2J")

router := proxy.NewRouter(store)

addr := "127.0.0.1:9090"
log.Printf("key-wallet daemon operational on %s (%d keys active in RAM)\n", addr, keysLoaded)

server := &http.Server{
Addr:    addr,
Handler: router,
}

if err := server.ListenAndServe(); err != nil {
log.Fatalf("Daemon failure: %v", err)
}
}

func min(a, b int) int {
if a < b {
return a
}
return b
}
