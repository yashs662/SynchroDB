package config

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/yashs662/SynchroDB/internal/logger"
	"github.com/yashs662/SynchroDB/internal/stores"
)

type Config struct {
	Debug              bool
	ID                 string
	Peers              []string
	Port               string
	JwtSecret          string
	Credentials        stores.CredentialStore
	CredentialFilePath string
	EncryptionKey      []byte
}

func (c *Config) validateConfig() {
	if c.ID == "" {
		logger.Error("Node ID cannot be empty")
		os.Exit(1)
	}

	if c.Port == "" {
		logger.Error("Port cannot be empty")
		os.Exit(1)
	}

	portNum, err := strconv.Atoi(c.Port)
	if err != nil || portNum < 1 || portNum > 65535 {
		logger.Error("Invalid port number. Must be between 1 and 65535")
		os.Exit(1)
	}

	if len(c.Peers) > 0 {
		logger.Infof("Peers: %v", c.Peers)
		logger.Infof("Number of peers: %d", len(c.Peers))
		for _, peer := range c.Peers {
			if peer == "" {
				logger.Error("Peer address cannot be empty")
				os.Exit(1)
			}
		}
	}
}

func (c *Config) ValidateEnvironment(setup *bool) {

	logger.Info("Running Environment checks...")

	if c.JwtSecret == "" {
		logger.Errorf("SYNCHRODB_JWT_SECRET environment variable not set")
		os.Exit(1)
	}

	if len(c.JwtSecret) < 32 {
		logger.Errorf("SYNCHRODB_JWT_SECRET must be at least 32 characters long")
		os.Exit(1)
	}

	if c.CredentialFilePath == "" {
		logger.Errorf("SYNCHRODB_CREDENTIAL_FILE_PATH environment variable not set")
		os.Exit(1)
	}

	if !*setup {
		if _, err := os.Stat(c.CredentialFilePath); os.IsNotExist(err) {
			logger.Errorf("Credential file does not exist at %s", c.CredentialFilePath)
			logger.Warn("Please run the setup command first")
			os.Exit(1)
		}
	}

	if c.EncryptionKey == nil {
		logger.Errorf("SYNCHRODB_ENCRYPTION_KEY environment variable not set")
		os.Exit(1)
	} else {
		if len(c.EncryptionKey) != 32 {
			logger.Error("Invalid encryption key length. Must be 32 bytes")
			os.Exit(1)
		}
	}

	logger.Info("Environment checks passed")
}

func (c *Config) loadEnvOverrides() {
	if val, exists := os.LookupEnv("SYNCHRODB_PORT"); exists {
		c.Port = val
	}
	if val, exists := os.LookupEnv("SYNCHRODB_ID"); exists {
		c.ID = val
	}
	// Add more overrides as needed
}

func ParseFlags() Config {
	debug := flag.Bool("debug", false, "enable debug mode with detailed logging")
	id := flag.String("id", "node1", "Node ID")
	peers := flag.String("peers", "", "Comma-separated list of peer addresses")
	port := flag.String("port", "8001", "Port to run the server on")
	registerUser := flag.Bool("register", false, "Register a new user")
	registerUsername := flag.String("username", "", "Username for registration")
	registerPassword := flag.String("password", "", "Password for registration")
	registerRole := flag.String("role", "admin", "Role for registration (admin, read-only, write-only, read-and-write)")
	setup := flag.Bool("setup", false, "Setup a new SynchroDB node")
	reset := flag.Bool("reset", false, "Reset the credential store")
	flag.BoolVar(debug, "d", false, "enable debug mode with detailed logging (shorthand)")
	flag.StringVar(id, "i", "node1", "Node ID (shorthand)")
	flag.StringVar(peers, "p", "", "Comma-separated list of peer addresses (shorthand)")
	flag.StringVar(port, "P", "8001", "Port to run the server on (shorthand)")
	flag.BoolVar(registerUser, "R", false, "Register a new user (shorthand)")
	flag.StringVar(registerUsername, "ru", "", "Username for registration (shorthand)")
	flag.StringVar(registerPassword, "rp", "", "Password for registration (shorthand)")
	flag.StringVar(registerRole, "rr", "admin", "Role for registration (shorthand)")
	flag.BoolVar(setup, "s", false, "Setup a new SynchroDB node (shorthand)")
	flag.BoolVar(reset, "r", false, "Reset the credential store (shorthand)")
	flag.Parse()

	// Load environment variables from .env file
	godotenv.Load()

	jwtSecret := os.Getenv("SYNCHRODB_JWT_SECRET")
	credential_file_path := os.Getenv("SYNCHRODB_CREDENTIAL_FILE_PATH")
	encryptionKey := []byte(os.Getenv("SYNCHRODB_ENCRYPTION_KEY"))

	var peerList []string
	if *peers != "" {
		peerList = strings.Split(*peers, ",")
	}

	config := Config{
		Debug:              *debug,
		ID:                 *id,
		Peers:              peerList,
		Port:               *port,
		JwtSecret:          jwtSecret,
		Credentials:        stores.CredentialStore{},
		CredentialFilePath: credential_file_path,
		EncryptionKey:      encryptionKey,
	}

	config.loadEnvOverrides()
	config.ValidateEnvironment(setup)

	// Setup a new SynchroDB node
	if *setup {
		// Check if any other flags are set
		flag.Visit(func(f *flag.Flag) {
			if f.Name != "debug" && f.Name != "d" && f.Name != "setup" && f.Name != "s" {
				logger.Warnf("Flag -%s is ignored when -setup is used", f.Name)
			}
		})
		Setup(config)
	}

	// Reset the credential store
	if *reset {
		ResetCredentials(config)
	}

	// Register a new user
	if *registerUser {
		role, err := parseRole(*registerRole)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
		RegisterUser(registerUsername, registerPassword, role, config.EncryptionKey, config.CredentialFilePath)
	} else if *registerUsername != "" || *registerPassword != "" {
		logger.Error("-ru and -rp flags must be used with -r flag")
		os.Exit(1)
	}

	config.validateConfig()

	logger.SetDebugMode(config.Debug)

	if config.Debug {
		logger.Debug("Debug mode enabled")
	}

	logger.Infof("Node ID: %s", config.ID)
	logger.Infof("Port: %s", config.Port)
	if len(config.Peers) > 0 {
		logger.Infof("Peers: %v", config.Peers)
	} else {
		logger.Warn("No peers provided")
	}

	return config
}

func parseRole(role string) (stores.Role, error) {
	switch role {
	case "admin":
		return stores.Admin, nil
	case "read-only":
		return stores.ReadOnly, nil
	case "write-only":
		return stores.WriteOnly, nil
	case "read-and-write":
		return stores.ReadAndWrite, nil
	default:
		return stores.Admin, fmt.Errorf("invalid role: %s", role)
	}
}

func Setup(config Config) {
	if _, err := os.Stat(config.CredentialFilePath); err == nil {
		logger.Error("Credential file already exists. Please remove the existing file or choose a different path")
		logger.Warn("Setup might've already been run, please check the credential file")
		os.Exit(1)
	} else {
		logger.Infof("Creating new credential file at %s", config.CredentialFilePath)

		err := os.MkdirAll(filepath.Dir(config.CredentialFilePath), os.ModePerm)
		if err != nil {
			logger.Errorf("Error creating directories: %v", err)
			os.Exit(1)
		}

		file, err := os.Create(config.CredentialFilePath)
		if err != nil {
			logger.Errorf("Error creating file: %v", err)
			os.Exit(1)
		}
		file.Close()
	}

	emptyCredentials := stores.CredentialStore{}

	err := emptyCredentials.AddUser("admin", "password", stores.Admin)
	if err != nil {
		logger.Errorf("Error adding user: %v", err)
		os.Exit(1)
	}

	err = stores.SaveCredentials(&emptyCredentials, config.EncryptionKey, config.CredentialFilePath)
	if err != nil {
		logger.Errorf("Error saving credentials: %v", err)
		os.Exit(1)
	}

	logger.Info("SynchroDB node setup complete")
	logger.Warn("Please change the default password for the admin user, it is recommended to use a strong password")
	os.Exit(0)
}

func RegisterUser(registerUsername *string, registerPassword *string, role stores.Role, encryptionKey []byte, credential_file_path string) {
	if *registerUsername == "" || *registerPassword == "" {
		logger.Error("Username and password must be provided for registration")
		os.Exit(1)
	}

	loadedCredentials, err := stores.LoadCredentials(encryptionKey, credential_file_path)
	if err != nil {
		logger.Errorf("Error loading credentials: %v", err)
		os.Exit(1)
	}

	err = loadedCredentials.AddUser(*registerUsername, *registerPassword, role)
	if err != nil {
		logger.Errorf("Error adding user: %v", err)
		os.Exit(1)
	}

	err = stores.SaveCredentials(loadedCredentials, encryptionKey, credential_file_path)
	if err != nil {
		logger.Errorf("Error saving credentials: %v", err)
		os.Exit(1)
	}

	logger.Infof("User %s registered successfully with role %s", *registerUsername, role.String())
	os.Exit(0)
}

func ResetCredentials(config Config) {
	reader := bufio.NewReader(os.Stdin)
	logger.Warn("Are you sure you want to reset the credential store? This action cannot be undone. (yes/no): ")
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)

	if response != "yes" && response != "y" {
		logger.Info("Reset action aborted by user")
		os.Exit(0)
	}

	logger.Warn("Resetting the credential store...")

	emptyCredentials := stores.CredentialStore{}

	err := emptyCredentials.AddUser("admin", "password", stores.Admin)
	if err != nil {
		logger.Errorf("Error adding user: %v", err)
		os.Exit(1)
	}

	err = stores.SaveCredentials(&emptyCredentials, config.EncryptionKey, config.CredentialFilePath)
	if err != nil {
		logger.Errorf("Error saving credentials: %v", err)
		os.Exit(1)
	}

	logger.Info("Credential store reset complete")
	logger.Warn("Please change the default password for the admin user, it is recommended to use a strong password")
	os.Exit(0)
}
