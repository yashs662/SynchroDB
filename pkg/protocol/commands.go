package protocol

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/yashs662/SynchroDB/internal/utils"
	"github.com/yashs662/SynchroDB/pkg/database"
)

func AllCommands(server *Server) []Command {
	return []Command{
		&AuthCommand{server: server},
		&PingCommand{server: server},
		&SetCommand{server: server},
		&GetCommand{server: server},
		&DelCommand{server: server},
		&ExpireCommand{server: server},
		&TTLCommand{server: server},
		&FlushDBCommand{server: server},
		&KeysCommand{server: server},
		&IncrCommand{server: server},
		&DecrCommand{server: server},
		&HelpCommand{server: server},
	}
}

type CommandDescription struct {
	Command  string
	Name     string
	Syntax   string
	HelpText string
}

type Command interface {
	Execute(conn net.Conn, args []string) string
	Replay(args []string, store *database.KVStore) error
	GetCommandInfo() CommandDescription
}

type AuthCommand struct {
	server *Server
}

func (c *AuthCommand) Execute(conn net.Conn, args []string) string {
	if len(args) != 1 {
		return "ERR missing password"
	}

	if args[0] == c.server.dbPassword {
		c.server.authenticateClient(conn)
		return "OK"
	}
	return "ERR invalid password"
}

func (c *AuthCommand) Replay(args []string, store *database.KVStore) error {
	return nil // No-op for replay
}

func (c *AuthCommand) GetCommandInfo() CommandDescription {
	return CommandDescription{
		Command:  "AUTH",
		Name:     "Authenticate",
		Syntax:   "AUTH <password>",
		HelpText: "Authenticate with the server",
	}
}

type PingCommand struct {
	// not needed here but kept for consistency
	server *Server
}

func (c *PingCommand) Execute(conn net.Conn, args []string) string {
	return "PONG"
}

func (c *PingCommand) Replay(args []string, store *database.KVStore) error {
	return nil // No-op for replay
}

func (c *PingCommand) GetCommandInfo() CommandDescription {
	return CommandDescription{
		Command:  "PING",
		Name:     "Ping",
		Syntax:   "PING",
		HelpText: "Check if the server is alive",
	}
}

type SetCommand struct {
	server *Server
}

func (c *SetCommand) Execute(conn net.Conn, args []string) string {
	if len(args) < 2 {
		return "ERR wrong number of arguments for 'SET' command"
	}
	key, value := args[0], args[1]
	if len(args) == 4 && args[2] == "EX" {
		ttl, err := strconv.Atoi(args[3])
		if err != nil || ttl <= 0 {
			return "ERR invalid TTL"
		}
		c.server.store.SetWithTTL(key, value, time.Duration(ttl)*time.Second)
		if c.server.persistenceEnabled && !strings.HasPrefix(key, "synchrodb-benchmark:") {
			c.server.aofWriter.Write(fmt.Sprintf("SET %s %s EX %d", key, value, ttl))
		}
		return "OK"
	} else if len(args) == 2 {
		c.server.store.Set(key, value)
		if c.server.persistenceEnabled && !strings.HasPrefix(key, "synchrodb-benchmark:") {
			c.server.aofWriter.Write(fmt.Sprintf("SET %s %s", key, value))
		}
		return "OK"
	}
	return "ERR invalid arguments for 'SET' command"
}

func (c *SetCommand) Replay(args []string, store *database.KVStore) error {
	if len(args) == 2 {
		key, value := args[0], args[1]
		store.Set(key, value)
	} else if len(args) == 4 && args[2] == "EX" {
		key, value := args[0], args[1]
		ttl, err := strconv.Atoi(args[3])
		if err != nil {
			return fmt.Errorf("invalid TTL value: %v", err)
		}
		store.SetWithTTL(key, value, time.Duration(ttl)*time.Second)
	} else {
		return fmt.Errorf("invalid arguments for 'SET' command")
	}
	return nil
}

func (c *SetCommand) GetCommandInfo() CommandDescription {
	return CommandDescription{
		Command:  "SET",
		Name:     "Set",
		Syntax:   "SET <key> <value> [EX <seconds>]",
		HelpText: "Set a key with a value and an optional expiration",
	}
}

type GetCommand struct {
	server *Server
}

func (c *GetCommand) Execute(conn net.Conn, args []string) string {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'GET' command"
	}
	key := args[0]
	value, exists := c.server.store.Get(key)
	if !exists {
		return "nil"
	}
	return value
}

func (c *GetCommand) Replay(args []string, store *database.KVStore) error {
	return nil // No-op for replay
}

func (c *GetCommand) GetCommandInfo() CommandDescription {
	return CommandDescription{
		Command:  "GET",
		Name:     "Get",
		Syntax:   "GET <key>",
		HelpText: "Get the value of a key",
	}
}

type DelCommand struct {
	server *Server
}

func (c *DelCommand) Execute(conn net.Conn, args []string) string {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'DEL' command"
	}
	key := args[0]
	if c.server.store.Del(key) {
		if c.server.persistenceEnabled && !strings.HasPrefix(key, "synchrodb-benchmark:") {
			c.server.aofWriter.Write(fmt.Sprintf("DEL %s", key))
		}
		return "OK"
	}
	return "nil"
}

func (c *DelCommand) Replay(args []string, store *database.KVStore) error {
	if len(args) != 1 {
		return fmt.Errorf("invalid arguments for 'DEL' command")
	}
	key := args[0]
	store.Del(key)
	return nil
}

func (c *DelCommand) GetCommandInfo() CommandDescription {
	return CommandDescription{
		Command:  "DEL",
		Name:     "Delete",
		Syntax:   "DEL <key>",
		HelpText: "Delete a key",
	}
}

type ExpireCommand struct {
	server *Server
}

func (c *ExpireCommand) Execute(conn net.Conn, args []string) string {
	if len(args) != 2 {
		return "ERR wrong number of arguments for 'EXPIRE' command"
	}
	key := args[0]
	ttl, err := strconv.Atoi(args[1])
	if err != nil || ttl <= 0 {
		return "ERR invalid TTL"
	}
	if c.server.store.SetExpire(key, ttl) {
		if c.server.persistenceEnabled && !strings.HasPrefix(key, "synchrodb-benchmark:") {
			c.server.aofWriter.Write(fmt.Sprintf("EXPIRE %s %d", key, ttl))
		}
		return "OK"
	}
	return "ERR key does not exist"
}

func (c *ExpireCommand) Replay(args []string, store *database.KVStore) error {
	if len(args) != 2 {
		return fmt.Errorf("invalid arguments for 'EXPIRE' command")
	}
	key := args[0]
	ttl, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid TTL value: %v", err)
	}
	store.SetExpire(key, ttl)
	return nil
}

func (c *ExpireCommand) GetCommandInfo() CommandDescription {
	return CommandDescription{
		Command:  "EXPIRE",
		Name:     "Expire",
		Syntax:   "EXPIRE <key> <seconds>",
		HelpText: "Set a key's time to live in seconds",
	}
}

type TTLCommand struct {
	server *Server
}

func (c *TTLCommand) Execute(conn net.Conn, args []string) string {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'TTL' command"
	}
	key := args[0]
	ttl := c.server.store.TTL(key)
	if ttl == -2 {
		return "-2"
	} else if ttl == -1 {
		return "-1"
	} else {
		return fmt.Sprintf("%ds", ttl)
	}
}

func (c *TTLCommand) Replay(args []string, store *database.KVStore) error {
	return nil // No-op for replay
}

func (c *TTLCommand) GetCommandInfo() CommandDescription {
	return CommandDescription{
		Command:  "TTL",
		Name:     "Time to Live",
		Syntax:   "TTL <key>",
		HelpText: "Get the time to live of a key",
	}
}

type FlushDBCommand struct {
	server *Server
}

func (c *FlushDBCommand) Execute(conn net.Conn, args []string) string {
	c.server.store.FlushDB()
	if c.server.persistenceEnabled {
		c.server.aofWriter.Write("FLUSHDB")
	}
	return "OK"
}

func (c *FlushDBCommand) Replay(args []string, store *database.KVStore) error {
	store.FlushDB()
	return nil
}

func (c *FlushDBCommand) GetCommandInfo() CommandDescription {
	return CommandDescription{
		Command:  "FLUSHDB",
		Name:     "Flush Database",
		Syntax:   "FLUSHDB",
		HelpText: "Remove all keys from the database",
	}
}

type KeysCommand struct {
	server *Server
}

func (c *KeysCommand) Execute(conn net.Conn, args []string) string {
	if len(args) < 1 {
		return "ERR missing pattern"
	}
	pattern := args[0]
	keys := c.server.store.Keys(pattern)
	if len(keys) > 20 {
		conn.Write([]byte("WARNING: More than 20 keys detected, displaying first 20 keys only.\n"))
		keys = keys[:20]
	}
	return strings.Join(keys, ", ")
}

func (c *KeysCommand) Replay(args []string, store *database.KVStore) error {
	return nil // No-op for replay
}

func (c *KeysCommand) GetCommandInfo() CommandDescription {
	return CommandDescription{
		Command:  "KEYS",
		Name:     "Keys",
		Syntax:   "KEYS <pattern>",
		HelpText: "Find all keys matching the given pattern",
	}
}

type IncrCommand struct {
	server *Server
}

func (c *IncrCommand) Execute(conn net.Conn, args []string) string {
	if len(args) < 1 {
		return "ERR missing key"
	}
	key := args[0]
	value, err := c.server.store.Incr(key)
	if err != nil {
		return fmt.Sprintf("ERR %v", err)
	}
	if c.server.persistenceEnabled && !strings.HasPrefix(key, "synchrodb-benchmark:") {
		c.server.aofWriter.Write(fmt.Sprintf("INCR %s", key))
	}
	return strconv.Itoa(value)
}

func (c *IncrCommand) Replay(args []string, store *database.KVStore) error {
	if len(args) != 1 {
		return fmt.Errorf("invalid arguments for 'INCR' command")
	}
	key := args[0]
	_, err := store.Incr(key)
	return err
}

func (c *IncrCommand) GetCommandInfo() CommandDescription {
	return CommandDescription{
		Command:  "INCR",
		Name:     "Increment",
		Syntax:   "INCR <key>",
		HelpText: "Increment the integer value of a key by one",
	}
}

type DecrCommand struct {
	server *Server
}

func (c *DecrCommand) Execute(conn net.Conn, args []string) string {
	if len(args) < 1 {
		return "ERR missing key"
	}
	key := args[0]
	value, err := c.server.store.Decr(key)
	if err != nil {
		return fmt.Sprintf("ERR %v", err)
	}
	if c.server.persistenceEnabled && !strings.HasPrefix(key, "synchrodb-benchmark:") {
		c.server.aofWriter.Write(fmt.Sprintf("DECR %s", key))
	}
	return strconv.Itoa(value)
}

func (c *DecrCommand) Replay(args []string, store *database.KVStore) error {
	if len(args) != 1 {
		return fmt.Errorf("invalid arguments for 'DECR' command")
	}
	key := args[0]
	_, err := store.Decr(key)
	return err
}

func (c *DecrCommand) GetCommandInfo() CommandDescription {
	return CommandDescription{
		Command:  "DECR",
		Name:     "Decrement",
		Syntax:   "DECR <key>",
		HelpText: "Decrement the integer value of a key by one",
	}
}

type HelpCommand struct {
	server *Server
}

func (c *HelpCommand) Execute(conn net.Conn, args []string) string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetHeader([]string{"Command", "Syntax", "Description"})
	table.SetRowLine(true)

	for _, command := range AllCommands(c.server) {
		description := command.GetCommandInfo()
		table.Append([]string{
			description.Name,
			description.Syntax,
			description.HelpText,
		})
	}

	table.Render()

	// replace the \n in the rendered string with <br> for the client to parse
	renderedString := utils.FormatMultilineResponse(tableString.String())

	return renderedString
}

func (c *HelpCommand) Replay(args []string, store *database.KVStore) error {
	return nil // No-op for replay
}

func (c *HelpCommand) GetCommandInfo() CommandDescription {
	return CommandDescription{
		Command:  "HELP",
		Name:     "Help",
		Syntax:   "HELP",
		HelpText: "Show this help message",
	}
}
