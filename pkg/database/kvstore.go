package database

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yashs662/SynchroDB/internal/logger"
)

type KVStore struct {
	data        map[string]string
	expirations map[string]time.Time
	mu          sync.RWMutex
}

func NewKVStore() *KVStore {
	store := &KVStore{
		data:        make(map[string]string),
		expirations: make(map[string]time.Time),
	}
	go store.cleanupExpiredKeys()
	return store
}

func (store *KVStore) Set(key, value string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.data[key] = value
	delete(store.expirations, key) // Remove expiration if it exists
}

func (store *KVStore) SetWithTTL(key, value string, ttl time.Duration) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.data[key] = value
	store.expirations[key] = time.Now().Add(ttl)
}

func (store *KVStore) SetExpire(key string, ttl int) bool {
	store.mu.Lock()
	defer store.mu.Unlock()
	if _, exists := store.data[key]; exists {
		store.expirations[key] = time.Now().Add(time.Duration(ttl) * time.Second)
		return true
	}
	return false
}

func (store *KVStore) Get(key string) (string, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	if exp, exists := store.expirations[key]; exists && time.Now().After(exp) {
		return "", false // Key has expired
	}

	value, exists := store.data[key]
	return value, exists
}

func (store *KVStore) TTL(key string) int {
	store.mu.RLock()
	defer store.mu.RUnlock()

	if exp, exists := store.expirations[key]; exists {
		if time.Now().After(exp) {
			return -2 // Key has expired
		}
		return int(time.Until(exp).Seconds())
	}

	if _, exists := store.data[key]; !exists {
		return -2 // Key does not exist
	} else {
		return -1 // Key exists but has no associated expiration
	}
}

func (store *KVStore) Del(key string) bool {
	store.mu.Lock()
	defer store.mu.Unlock()
	if _, exists := store.data[key]; exists {
		delete(store.data, key)
		delete(store.expirations, key)
		return true
	}
	return false
}

func (store *KVStore) cleanupExpiredKeys() {
	for {
		time.Sleep(1 * time.Second) // Run every second
		now := time.Now()
		store.mu.Lock()
		for key, exp := range store.expirations {
			if now.After(exp) {
				delete(store.data, key)
				delete(store.expirations, key)
			}
		}
		store.mu.Unlock()
	}
}

func (store *KVStore) LoadFromAOF(filepath string) error {
	logger.Infof("Replaying AOF file: %s", filepath)
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			logger.Debugf("Skipping Malformed line in AOF: %s", line)
			continue // Malformed line
		}

		// Parse timestamp and command
		timestamp, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			logger.Debugf("Skipping Invalid timestamp in AOF: %s", line)
			continue // Skip invalid entries
		}
		command := parts[1]
		args := parts[2:]

		// Check if the command is still valid
		if command == "SET" {
			if len(args) == 2 {
				key, value := args[0], args[1]
				store.Set(key, value)
			} else if len(args) == 4 && args[2] == "EX" {
				key, value := args[0], args[1]
				ttl, err := strconv.Atoi(args[3])
				if err == nil {
					store.SetWithTTL(key, value, time.Duration(ttl)*time.Second)
				} else {
					logger.Warnf("Invalid TTL value in AOF: %s", line)
					logger.Warn("Ignoring the command")
				}
			} else {
				logger.Warnf("Invalid SET command in AOF: %s", line)
				logger.Warn("Ignoring the command")
			}
		} else if command == "EXPIRE" {
			if len(args) != 2 {
				logger.Warnf("Invalid EXPIRE command in AOF: %s", line)
				logger.Warn("Ignoring the command")
				continue
			}

			key := args[0]
			ttl, err := strconv.Atoi(args[1])
			if err == nil {
				expiration := time.Unix(timestamp, 0).Add(time.Duration(ttl) * time.Second)
				if time.Now().Before(expiration) {
					value, exists := store.data[key]
					if !exists {
						logger.Warnf("Key %s does not exist for EXPIRE command in AOF: %s", key, line)
						logger.Warn("Ignoring the command")
						continue
					}
					remainingTTL := time.Until(expiration)
					store.SetWithTTL(key, value, remainingTTL)
				}
			}
		} else if command == "DEL" {
			if len(args) != 1 {
				logger.Warnf("Invalid DEL command in AOF: %s", line)
				logger.Warn("Ignoring the command")
				continue
			}
			store.Del(args[0])
		}
	}

	logger.Info("AOF replay completed")

	return scanner.Err()
}
