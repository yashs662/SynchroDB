package database

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yashs662/SynchroDB/internal/logger"
	"github.com/yashs662/SynchroDB/internal/utils"
)

type Command interface {
	Execute(conn net.Conn, args []string) string
	Replay(args []string, store *KVStore) error
}
type CommandRegistry struct {
	commands map[string]Command
}

func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{commands: make(map[string]Command)}
}

func (r *CommandRegistry) Register(name string, cmd Command) {
	r.commands[strings.ToUpper(name)] = cmd
}

func (r *CommandRegistry) Get(name string) (Command, bool) {
	cmd, exists := r.commands[strings.ToUpper(name)]
	return cmd, exists
}

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

func (store *KVStore) LoadFromAOF(filepath string, commandRegistry *CommandRegistry) error {
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
		_, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			logger.Debugf("Skipping Invalid timestamp in AOF: %s", line)
			continue // Skip invalid entries
		}
		command := parts[1]
		args := parts[2:]

		// Get the command from the registry and replay it
		cmd, exists := commandRegistry.Get(command)
		if !exists {
			logger.Warnf("Unknown command in AOF: %s", line)
			continue
		}

		if err := cmd.Replay(args, store); err != nil {
			logger.Warnf("Failed to replay command in AOF: %s, error: %v", line, err)
		}
	}

	logger.Info("AOF replay completed")

	return scanner.Err()
}

func (store *KVStore) FlushDB() {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.data = make(map[string]string)
	store.expirations = make(map[string]time.Time)
}

func (store *KVStore) Keys(pattern string) []string {
	store.mu.RLock()
	defer store.mu.RUnlock()
	keys := []string{}
	for key := range store.data {
		if utils.MatchPattern(key, pattern) {
			keys = append(keys, key)
		}
	}
	return keys
}

func (store *KVStore) Incr(key string) (int, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	value, exists := store.data[key]
	if !exists {
		store.data[key] = "1"
		return 1, nil
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("value is not an integer")
	}
	intValue++
	store.data[key] = strconv.Itoa(intValue)
	return intValue, nil
}

func (store *KVStore) Decr(key string) (int, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	value, exists := store.data[key]
	if !exists {
		store.data[key] = "-1"
		return -1, nil
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("value is not an integer")
	}
	intValue--
	store.data[key] = strconv.Itoa(intValue)
	return intValue, nil
}
