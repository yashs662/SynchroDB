package stores

import (
	"sync"
)

type KVStore struct {
	mu          sync.RWMutex
	store       sync.Map
	Credentials CredentialStore
}

func NewStore() *KVStore {
	return &KVStore{}
}

func (s *KVStore) Get(key string) (string, bool) {
	value, exists := s.store.Load(key)
	if !exists {
		return "", false
	}
	return value.(string), true
}

func (s *KVStore) Set(key, value string) {
	s.store.Store(key, value)
}

func (s *KVStore) Delete(key string) {
	s.store.Delete(key)
}

func (s *KVStore) GetAllKeys() []string {
	keys := []string{}
	s.store.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})
	return keys
}
