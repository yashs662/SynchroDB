package protocol

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/yashs662/SynchroDB/internal/config"
	"github.com/yashs662/SynchroDB/internal/logger"
	"github.com/yashs662/SynchroDB/pkg/database"
)

var (
	listener             net.Listener
	conns                sync.Map
	authenticatedClients = make(map[net.Conn]bool)
	authEnabled          bool
	dbPassword           string
	store                = database.NewKVStore()
)

func StartServer(config *config.Config) error {

	authEnabled = config.Server.AuthEnabled
	dbPassword = config.Server.Password

	cert, err := tls.LoadX509KeyPair("server-cert.pem", "server-key.pem")
	if err != nil {
		return fmt.Errorf("failed to load TLS certificates: %w", err)
	}

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
	listener, err := tls.Listen("tcp", config.Server.Address, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to start TLS listener: %w", err)
	}
	defer listener.Close()

	logger.Infof("Secure server is listening on %s", config.Server.Address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Warnf("Failed to accept connection: %v", err)
			continue
		}

		conns.Store(conn, struct{}{})
		go handleConnection(conn)
	}
}

func Shutdown(ctx context.Context) error {
	if listener != nil {
		listener.Close()
	}

	var wg sync.WaitGroup
	conns.Range(func(key, value interface{}) bool {
		conn := key.(net.Conn)
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn.Close()
		}()
		return true
	})

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		conns.Delete(conn)
	}()
	clientAddr := conn.RemoteAddr().String()
	logger.Infof("Accepted connection from %s", clientAddr)

	reader := bufio.NewReader(conn)

	for {
		command, err := reader.ReadString('\n')
		if err != nil {
			logger.Warnf("Connection closed by %s: %v", clientAddr, err)
			return
		}

		command = strings.TrimSpace(command)
		response := handleCommand(conn, command)
		conn.Write([]byte(response + "\n"))
	}
}

func handleCommand(conn net.Conn, command string) string {
	// Enforce authentication
	if authEnabled {
		if !authenticatedClients[conn] && !strings.HasPrefix(command, "AUTH ") {
			return "ERR authentication required"
		}
	}

	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "ERR invalid command"
	}

	switch parts[0] {
	case "AUTH":
		return handleAuth(conn, parts[1:])
	case "PING":
		return "PONG"
	case "SET":
		return handleSet(parts[1:])
	case "GET":
		return handleGet(parts[1:])
	case "DEL":
		return handleDel(parts[1:])
	default:
		return "ERR unknown command"
	}
}

func handleSet(args []string) string {
	if len(args) != 2 {
		return "ERR wrong number of arguments for 'SET' command"
	}
	key, value := args[0], args[1]
	store.Set(key, value)
	return "OK"
}

func handleGet(args []string) string {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'GET' command"
	}
	key := args[0]
	value, exists := store.Get(key)
	if !exists {
		return "nil"
	}
	return value
}

func handleDel(args []string) string {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'DEL' command"
	}
	key := args[0]
	if store.Del(key) {
		return "OK"
	}
	return "nil"
}

func handleAuth(conn net.Conn, args []string) string {
	if len(args) != 1 {
		return "ERR missing password"
	}

	if args[0] == dbPassword {
		authenticatedClients[conn] = true
		return "OK"
	}
	return "ERR invalid password"
}
