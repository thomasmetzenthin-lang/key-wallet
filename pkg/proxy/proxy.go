package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"key-wallet/pkg/keyring"
)

type ProviderConfig struct {
	TargetURL *url.URL
}

type Server struct {
	store     keyring.KeyStore
	providers map[string]*ProviderConfig
}

func NewServer(store keyring.KeyStore) *Server {
	openRouterURL, _ := url.Parse("https://openrouter.ai/api/v1")
	openAIURL, _ := url.Parse("https://api.openai.com/v1")

	return &Server{
		store: store,
		providers: map[string]*ProviderConfig{
			"openrouter": {TargetURL: openRouterURL},
			"openai":     {TargetURL: openAIURL},
		},
	}
}

func (s *Server) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Restrict origin strictly to local loopback interface
		host := strings.Split(r.RemoteAddr, ":")[0]
		if host != "127.0.0.1" && host != "[::1]" && host != "localhost" {
			http.Error(w, "Forbidden: Access limited to local loopback interface.", http.StatusForbidden)
			return
		}

		// 2. Route Parsing: Expected path format /proxy/{provider}/...
		trimmedPath := strings.TrimPrefix(r.URL.Path, "/")
		parts := strings.Split(trimmedPath, "/")

		if len(parts) < 2 || parts[0] != "proxy" {
			http.Error(w, "Invalid route. Syntax: /proxy/{provider}/endpoint", http.StatusBadRequest)
			return
		}

		providerKey := parts[1]
		provider, exists := s.providers[providerKey]
		if !exists {
			http.Error(w, fmt.Sprintf("Unsupported provider: %s", providerKey), http.StatusNotFound)
			return
		}

		// 3. Resolve key from storage
		apiKey, err := s.store.Get(providerKey)
		if err != nil {
			http.Error(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusUnauthorized)
			return
		}

		// 4. Construct Reverse Proxy with dynamic header mutation
		proxy := httputil.NewSingleHostReverseProxy(provider.TargetURL)

		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)

			// Rewrite target URL path
			req.Host = provider.TargetURL.Host
			req.URL.Path = "/" + strings.Join(parts[2:], "/")

			// Inject Bearer credential header on volatile request flight
			req.Header.Set("Authorization", "Bearer "+apiKey)
			req.Header.Del("X-Forwarded-For") // Prevent local IP leakage
		}

		proxy.ServeHTTP(w, r)
	}
}
