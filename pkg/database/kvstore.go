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
	data        sync.Map
	expirations sync.Map
}

func NewKVStore() *KVStore {
	store := &KVStore{}
	go store.cleanupExpiredKeys()
	return store
}

func (store *KVStore) Set(key, value string) {
	store.data.Store(key, value)
	store.expirations.Delete(key) // Remove expiration if it exists
}

func (store *KVStore) SetWithTTL(key, value string, ttl time.Duration) {
	store.data.Store(key, value)
	store.expirations.Store(key, time.Now().Add(ttl))
}

func (store *KVStore) SetExpire(key string, ttl int) bool {
	if _, exists := store.data.Load(key); exists {
		store.expirations.Store(key, time.Now().Add(time.Duration(ttl)*time.Second))
		return true
	}
	return false
}

func (store *KVStore) Get(key string) (string, bool) {
	if exp, exists := store.expirations.Load(key); exists {
		if time.Now().After(exp.(time.Time)) {
			return "", false // Key has expired
		}
	}

	value, exists := store.data.Load(key)
	if !exists {
		return "", false
	}
	return value.(string), true
}

func (store *KVStore) TTL(key string) int {
	if exp, exists := store.expirations.Load(key); exists {
		if time.Now().After(exp.(time.Time)) {
			return -2 // Key has expired
		}
		return int(time.Until(exp.(time.Time)).Seconds())
	}

	if _, exists := store.data.Load(key); !exists {
		return -2 // Key does not exist
	} else {
		return -1 // Key exists but has no associated expiration
	}
}

func (store *KVStore) Del(key string) bool {
	if _, exists := store.data.Load(key); exists {
		store.data.Delete(key)
		store.expirations.Delete(key)
		return true
	}
	return false
}

func (store *KVStore) cleanupExpiredKeys() {
	for {
		time.Sleep(1 * time.Second) // Run every second
		now := time.Now()
		store.expirations.Range(func(key, exp interface{}) bool {
			if now.After(exp.(time.Time)) {
				store.data.Delete(key)
				store.expirations.Delete(key)
			}
			return true
		})
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
	store.data = sync.Map{}
	store.expirations = sync.Map{}
}

func (store *KVStore) Keys(pattern string) []string {
	keys := []string{}
	store.data.Range(func(key, value interface{}) bool {
		if utils.MatchPattern(key.(string), pattern) {
			keys = append(keys, key.(string))
		}
		return true
	})
	return keys
}

func (store *KVStore) Incr(key string) (int, error) {
	value, exists := store.data.Load(key)
	if !exists {
		store.data.Store(key, "1")
		return 1, nil
	}
	intValue, err := strconv.Atoi(value.(string))
	if err != nil {
		return 0, fmt.Errorf("value is not an integer")
	}
	intValue++
	store.data.Store(key, strconv.Itoa(intValue))
	return intValue, nil
}

func (store *KVStore) Decr(key string) (int, error) {
	value, exists := store.data.Load(key)
	if !exists {
		store.data.Store(key, "-1")
		return -1, nil
	}
	intValue, err := strconv.Atoi(value.(string))
	if err != nil {
		return 0, fmt.Errorf("value is not an integer")
	}
	intValue--
	store.data.Store(key, strconv.Itoa(intValue))
	return intValue, nil
}
