package proxy

import (
"bufio"
"encoding/json"
"fmt"
"io"
"net/http"
"net/http/httputil"
"net/url"
"strings"

"key-wallet/pkg/keyring"
"key-wallet/web"
)

type Router struct {
store     *keyring.SecureStore
providers map[string]string
}

func NewRouter(store *keyring.SecureStore) *Router {
return &Router{
store: store,
providers: map[string]string{
"openrouter": "https://openrouter.ai/api/v1",
"openai":     "https://api.openai.com/v1",
"anthropic":  "https://api.anthropic.com/v1",
"gemini":     "https://generativelanguage.googleapis.com/v1beta/openai",
},
}
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 1. Dashboard Root Handler
if r.URL.Path == "/" {
w.Header().Set("Content-Type", "text/html; charset=utf-8")
w.Write(web.IndexHTML)
return
}

// 2. Active Key Status API
if r.URL.Path == "/api/status" {
rt.handleStatus(w, r)
return
}

// 3. Key Loader API
if r.URL.Path == "/api/keys" {
rt.handleLoadKeys(w, r)
return
}

// 4. Upstream Reverse Proxy Router (/proxy/{provider}/...)
pathSegments := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
if len(pathSegments) < 2 || pathSegments[0] != "proxy" {
http.Error(w, "Malformed proxy route path. Expected /proxy/{provider}/...", http.StatusBadRequest)
return
}

provider := pathSegments[1]
targetBase, exists := rt.providers[provider]
if !exists {
http.Error(w, "Unrecognized target provider", http.StatusBadRequest)
return
}

targetURL, err := url.Parse(targetBase)
if err != nil {
http.Error(w, "Upstream endpoint parse failure", http.StatusInternalServerError)
return
}

rawKey, err := rt.store.GetDecrypt(provider)
if err != nil {
http.Error(w, "Credential store access error: key not loaded for provider "+provider, http.StatusUnauthorized)
return
}

// Clean trailing whitespace and control characters (\r, \n)
cleanKey := strings.TrimSpace(string(rawKey))
keyring.WipeByteSlice(rawKey)

proxy := httputil.NewSingleHostReverseProxy(targetURL)
originalDirector := proxy.Director

proxy.Director = func(req *http.Request) {
originalDirector(req)

subPath := "/" + strings.Join(pathSegments[2:], "/")
req.URL.Scheme = targetURL.Scheme
req.URL.Host = targetURL.Host
req.URL.Path = singleJoiningSlash(targetURL.Path, subPath)
req.Host = targetURL.Host

// Format auth header based on provider requirements
switch provider {
case "anthropic":
req.Header.Set("x-api-key", cleanKey)
req.Header.Set("anthropic-version", "2023-06-01")
case "gemini":
req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cleanKey))
req.Header.Set("x-goog-api-key", cleanKey)
default: // openrouter, openai
req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cleanKey))
}

req.Header.Del("X-Forwarded-For")
}

proxy.ServeHTTP(w, r)
}

func (rt *Router) handleStatus(w http.ResponseWriter, req *http.Request) {
if req.Method != http.MethodGet {
http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
return
}

providers := []string{"openrouter", "anthropic", "gemini", "openai"}
status := make(map[string]bool)

for _, p := range providers {
rawKey, err := rt.store.GetDecrypt(p)
status[p] = (err == nil)
if err == nil {
keyring.WipeByteSlice(rawKey)
}
}

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(status)
}

func (rt *Router) handleLoadKeys(w http.ResponseWriter, req *http.Request) {
if req.Method != http.MethodPost {
http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
return
}

body, err := io.ReadAll(req.Body)
if err != nil {
http.Error(w, "Failed to read request", http.StatusBadRequest)
return
}

scanner := bufio.NewScanner(strings.NewReader(string(body)))
keysLoaded := 0

for scanner.Scan() {
line := strings.TrimSpace(scanner.Text())
if line == "" || strings.HasPrefix(line, "#") {
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
continue
}

if err := rt.store.Set(provider, []byte(line)); err == nil {
keysLoaded++
}
}

w.Header().Set("Content-Type", "text/plain")
fmt.Fprintf(w, "[+] Loaded %d key(s) into RAM", keysLoaded)
}

func singleJoiningSlash(a, b string) string {
asSlash := strings.HasSuffix(a, "/")
bsSlash := strings.HasPrefix(b, "/")
switch {
case asSlash && bsSlash:
return a + b[1:]
case !asSlash && !bsSlash:
return a + "/" + b
}
return a + b
}
