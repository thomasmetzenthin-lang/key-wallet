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

initDesktopFeatures()

store, err := keyring.NewSecureStore()
if err != nil {
log.Fatalf("Failed to initialize secure store: %v", err)
}

// Launch background CLI key reader (non-blocking)
go runCLIKeyReader(store)

router := proxy.NewRouter(store)
addr := "127.0.0.1:9090"

log.Printf("key-wallet daemon operational on http://%s\n", addr)
log.Printf("Web Dashboard live at http://%s/\n", addr)

server := &http.Server{
Addr:    addr,
Handler: router,
}

if err := server.ListenAndServe(); err != nil {
log.Fatalf("Daemon failure: %v", err)
}
}

func runCLIKeyReader(store *keyring.SecureStore) {
fmt.Println("Paste keys below OR load via Dashboard at http://127.0.0.1:9090")
fmt.Println("Type 'DONE' or press Enter on an empty line when finished:\n")

scanner := bufio.NewScanner(os.Stdin)
keysLoaded := 0

for scanner.Scan() {
line := strings.TrimSpace(scanner.Text())

if line == "" || strings.ToUpper(line) == "DONE" {
if keysLoaded > 0 {
fmt.Printf("[+] Finished loading %d key(s) from CLI.\n", keysLoaded)
break
}
continue
}

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
fmt.Printf("[?] Unrecognized format for '%s...'. Skipping.\n", line[:min(len(line), 8)])
continue
}

if err := store.Set(provider, []byte(line)); err != nil {
fmt.Printf("[!] Error storing key for %s: %v\n", provider, err)
continue
}

fmt.Printf("[+] Loaded %s key into RAM.\n", provider)
keysLoaded++
}
}

func min(a, b int) int {
if a < b {
return a
}
return b
}
