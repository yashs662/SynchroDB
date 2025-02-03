package protocol

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/yashs662/SynchroDB/internal/config"
	"github.com/yashs662/SynchroDB/internal/logger"
	"github.com/yashs662/SynchroDB/pkg/database"
)

type Server struct {
	listener             net.Listener
	conns                sync.Map
	authenticatedClients map[net.Conn]bool
	authMutex            sync.RWMutex
	authEnabled          bool
	dbPassword           string
	store                *database.KVStore
	aofWriter            *database.AOFWriter
	persistenceEnabled   bool
	commandRegistry      *database.CommandRegistry
	connCount            int
	connMutex            sync.Mutex
	maxConnections       int
	rateLimit            int
	shutdownChan         chan struct{}
}

func NewServer(config *config.Config, store *database.KVStore, aofWriter *database.AOFWriter) *Server {
	server := &Server{
		authEnabled:          config.Server.AuthEnabled,
		dbPassword:           config.Server.Password,
		store:                store,
		aofWriter:            aofWriter,
		authenticatedClients: make(map[net.Conn]bool),
		commandRegistry:      database.NewCommandRegistry(),
		maxConnections:       config.Server.MaxConnections,
		rateLimit:            config.Server.RateLimit,
		shutdownChan:         make(chan struct{}),
	}

	// Register commands
	server.registerCommands()

	return server
}

func (s *Server) registerCommands() {
	commands := AllCommands(s)

	for _, cmd := range commands {
		s.commandRegistry.Register(cmd.GetCommandInfo().Command, cmd)
	}
}

func (s *Server) Start(config *config.Config) error {
	s.authEnabled = config.Server.AuthEnabled
	s.dbPassword = config.Server.Password
	aofFilePath := config.Server.PersistentAOFPath
	if aofFilePath != "" {
		var err error
		s.aofWriter, err = database.NewAOFWriter(aofFilePath)
		if err != nil {
			return fmt.Errorf("failed to create AOF writer: %w", err)
		}
		s.persistenceEnabled = true

		if config.Server.ReplayAOFOnStartup {
			file, err := os.Open(aofFilePath)
			if err != nil {
				logger.Warnf("Failed to open AOF file: %v", err)
			} else {
				defer file.Close()
				s.store.LoadFromAOF(aofFilePath, s.commandRegistry)
			}
		}

		defer s.aofWriter.Close()
	} else {
		logger.Warn("Persistence is disabled because the file path is empty in the config")
	}

	cert, err := tls.LoadX509KeyPair("server-cert.pem", "server-key.pem")
	if err != nil {
		return fmt.Errorf("failed to load TLS certificates: %w", err)
	}

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
	s.listener, err = tls.Listen("tcp", config.Server.Address, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to start TLS listener: %w", err)
	}
	defer s.listener.Close()

	logger.Infof("Secure server is listening on %s", config.Server.Address)

	for {
		conn, err := s.listener.Accept()
		select {
		case <-s.shutdownChan:
			return nil
		default:
		}
		if err != nil {
			logger.Warnf("Failed to accept connection: %v", err)
			continue
		}

		s.connMutex.Lock()
		if s.maxConnections > 0 && s.connCount >= s.maxConnections {
			s.connMutex.Unlock()
			logger.Warnf("Connection limit reached, rejecting connection from %s", conn.RemoteAddr().String())
			conn.Close()
			continue
		}
		s.connCount++
		s.connMutex.Unlock()

		logger.Debugf("Accepted connection from %s", conn.RemoteAddr().String())

		s.conns.Store(conn, struct{}{})
		go s.handleConnection(conn)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	close(s.shutdownChan)
	if s.listener != nil {
		s.listener.Close()
	}

	var wg sync.WaitGroup
	s.conns.Range(func(key, value interface{}) bool {
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

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.conns.Delete(conn)
		s.connMutex.Lock()
		s.connCount--
		s.connMutex.Unlock()
	}()
	clientAddr := conn.RemoteAddr().String()
	logger.Debugf("Handling connection from %s", clientAddr)

	tlsConn, ok := conn.(*tls.Conn)
	if ok {
		if err := tlsConn.Handshake(); err != nil {
			logger.Errorf("TLS handshake failed with %s: %v", clientAddr, err)
			return
		}
	}

	reader := bufio.NewReader(conn)
	var rateLimiter <-chan time.Time
	if s.rateLimit > 0 {
		rateLimiter = time.Tick(time.Second / time.Duration(s.rateLimit))
	}

	for {
		if s.rateLimit > 0 {
			<-rateLimiter
		}
		command, err := reader.ReadString('\n')
		if err != nil {
			logger.Debugf("Connection closed by %s: %v", clientAddr, err)
			return
		}

		command = strings.TrimSpace(command)
		response := s.handleCommand(conn, command)
		conn.Write([]byte(response + "\n"))
	}
}

func (s *Server) handleCommand(conn net.Conn, command string) string {
	// Enforce authentication
	if s.authEnabled {
		s.authMutex.RLock()
		authenticated := s.authenticatedClients[conn]
		s.authMutex.RUnlock()

		authCommand := AuthCommand{}
		authCommandName := authCommand.GetCommandInfo().Command
		// add a space after the command name to avoid partial matches
		// e.g. AUTH and AUTHENTICATE
		authCommandName += " "

		if !authenticated && !strings.HasPrefix(strings.ToUpper(command), authCommandName) {
			return "ERR authentication required"
		}
	}

	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "ERR invalid command"
	}

	cmd, exists := s.commandRegistry.Get(strings.ToUpper(parts[0]))
	if !exists {
		return "ERR unknown command"
	}

	return cmd.Execute(conn, parts[1:])
}

func (s *Server) authenticateClient(conn net.Conn) {
	s.authMutex.Lock()
	s.authenticatedClients[conn] = true
	s.authMutex.Unlock()
}
