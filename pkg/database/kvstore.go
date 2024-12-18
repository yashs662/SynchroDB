package database

import "sync"

type KVStore struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewKVStore() *KVStore {
	return &KVStore{
		data: make(map[string]string),
	}
}

func (store *KVStore) Set(key, value string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.data[key] = value
}

func (store *KVStore) Get(key string) (string, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	value, exists := store.data[key]
	return value, exists
}

func (store *KVStore) Del(key string) bool {
	store.mu.Lock()
	defer store.mu.Unlock()
	if _, exists := store.data[key]; exists {
		delete(store.data, key)
		return true
	}
	return false
}
