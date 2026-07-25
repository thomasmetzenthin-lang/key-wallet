package keyring

import "fmt"

type KeyStore interface {
	Get(provider string) (string, error)
	Set(provider, secret string) error
	Delete(provider string) error
}

type MemoryKeyStore struct {
	secrets map[string]string
}

func NewMemoryKeyStore() *MemoryKeyStore {
	return &MemoryKeyStore{
		secrets: make(map[string]string),
	}
}

func (m *MemoryKeyStore) Get(provider string) (string, error) {
	secret, exists := m.secrets[provider]
	if !exists {
		return "", fmt.Errorf("secret for provider '%s' not found", provider)
	}
	return secret, nil
}

func (m *MemoryKeyStore) Set(provider, secret string) error {
	m.secrets[provider] = secret
	return nil
}

func (m *MemoryKeyStore) Delete(provider string) error {
	delete(m.secrets, provider)
	return nil
}
