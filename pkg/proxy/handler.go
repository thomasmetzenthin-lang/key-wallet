package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"key-wallet/pkg/keyring"
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
	pathSegments := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(pathSegments) < 2 || pathSegments[0] != "proxy" {
		http.Error(w, "Malformed proxy route path", http.StatusBadRequest)
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
