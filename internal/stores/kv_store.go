package stores

import (
	"sync"
)

type KVStore struct {
	mu          sync.RWMutex
	store       map[string]string
	Credentials CredentialStore
}

func NewStore() *KVStore {
	return &KVStore{
		store: make(map[string]string),
	}
}

func (s *KVStore) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.store[key]
	return value, exists
}

func (s *KVStore) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[key] = value
}

func (s *KVStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, key)
}
