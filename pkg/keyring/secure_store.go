package keyring

import (
"crypto/aes"
"crypto/cipher"
"crypto/rand"
"errors"
"io"
"sync"
)

var (
ErrKeyNotFound = errors.New("keyring: requested key identifier non-existent")
ErrCorruptData = errors.New("keyring: ciphertext authentication or integrity failure")
)

type SecureStore struct {
mu        sync.RWMutex
masterKey []byte
store     map[string][]byte
}

func NewSecureStore() (*SecureStore, error) {
mk := make([]byte, 32)
if _, err := io.ReadFull(rand.Reader, mk); err != nil {
return nil, err
}

return &SecureStore{
masterKey: mk,
store:     make(map[string][]byte),
}, nil
}

func (s *SecureStore) Set(provider string, secret []byte) error {
s.mu.Lock()
defer s.mu.Unlock()

block, err := aes.NewCipher(s.masterKey)
if err != nil {
return err
}

gcm, err := cipher.NewGCM(block)
if err != nil {
return err
}

nonce := make([]byte, gcm.NonceSize())
if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
return err
}

ciphertext := gcm.Seal(nonce, nonce, secret, nil)
s.store[provider] = ciphertext
return nil
}

func (s *SecureStore) GetDecrypt(provider string) ([]byte, error) {
s.mu.RLock()
ciphertext, exists := s.store[provider]
s.mu.RUnlock()

if !exists {
return nil, ErrKeyNotFound
}

block, err := aes.NewCipher(s.masterKey)
if err != nil {
return nil, err
}

gcm, err := cipher.NewGCM(block)
if err != nil {
return nil, err
}

nonceSize := gcm.NonceSize()
if len(ciphertext) < nonceSize {
return nil, ErrCorruptData
}

nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
plaintext, err := gcm.Open(nil, nonce, actualCiphertext, nil)
if err != nil {
return nil, ErrCorruptData
}

return plaintext, nil
}

func WipeByteSlice(b []byte) {
for i := range b {
b[i] = 0
}
}

func (s *SecureStore) Close() {
s.mu.Lock()
defer s.mu.Unlock()

WipeByteSlice(s.masterKey)
for k, v := range s.store {
WipeByteSlice(v)
delete(s.store, k)
}
}
