package config

import (
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
		fmt.Println("Error: Node ID cannot be empty")
		os.Exit(1)
	}

	if c.Port == "" {
		fmt.Println("Error: Port cannot be empty")
		os.Exit(1)
	}

	portNum, err := strconv.Atoi(c.Port)
	if err != nil || portNum < 1 || portNum > 65535 {
		fmt.Println("Error: Invalid port number. Must be between 1 and 65535")
		os.Exit(1)
	}

	if len(c.Peers) > 0 {
		logger.Infof("Peers: %v", c.Peers)
		logger.Infof("Number of peers: %d", len(c.Peers))
		for _, peer := range c.Peers {
			if peer == "" {
				fmt.Println("Error: Peer address cannot be empty")
				os.Exit(1)
			}
		}
	}
}

func (c *Config) ValidateEnvironment() {

	logger.Info("Running Environment checks...")

	if c.JwtSecret == "" {
		logger.Errorf("SYNCHRODB_JWT_SECRET environment variable not set")
		os.Exit(1)
	}

	if c.CredentialFilePath == "" {
		logger.Errorf("SYNCHRODB_CREDENTIAL_FILE_PATH environment variable not set")
		os.Exit(1)
	}

	if c.EncryptionKey == nil {
		logger.Errorf("SYNCHRODB_ENCRYPTION_KEY environment variable not set")
		os.Exit(1)
	}

	logger.Info("Environment checks passed")
}

func ParseFlags() Config {
	debug := flag.Bool("debug", false, "enable debug mode with detailed logging")
	id := flag.String("id", "node1", "Node ID")
	peers := flag.String("peers", "", "Comma-separated list of peer addresses")
	port := flag.String("port", "8001", "Port to run the server on")
	registerUser := flag.Bool("register", false, "Register a new user")
	registerUsername := flag.String("username", "", "Username for registration")
	registerPassword := flag.String("password", "", "Password for registration")
	setup := flag.Bool("setup", false, "Setup a new SynchroDB node")
	flag.BoolVar(debug, "d", false, "enable debug mode with detailed logging (shorthand)")
	flag.StringVar(id, "i", "node1", "Node ID (shorthand)")
	flag.StringVar(peers, "p", "", "Comma-separated list of peer addresses (shorthand)")
	flag.StringVar(port, "P", "8001", "Port to run the server on (shorthand)")
	flag.BoolVar(registerUser, "r", false, "Register a new user (shorthand)")
	flag.StringVar(registerUsername, "ru", "", "Username for registration (shorthand)")
	flag.StringVar(registerPassword, "rp", "", "Password for registration (shorthand)")
	flag.BoolVar(setup, "s", false, "Setup a new SynchroDB node (shorthand)")
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

	config.ValidateEnvironment()

	// Setup a new SynchroDB node
	if *setup {
		Setup(config)
	}

	// Register a new user
	if *registerUser {
		RegisterUser(registerUsername, registerPassword, config.EncryptionKey, config.CredentialFilePath)
	} else if *registerUsername != "" || *registerPassword != "" {
		fmt.Println("Error: -ru and -rp flags must be used with -r flag")
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

func Setup(config Config) {
	if _, err := os.Stat(config.CredentialFilePath); err == nil {
		fmt.Println("Error: Credential file already exists. Please remove the existing file or choose a different path")
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

	err := emptyCredentials.AddUser("admin", "password", "admin")
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

func RegisterUser(registerUsername *string, registerPassword *string, encryptionKey []byte, credential_file_path string) {
	if *registerUsername == "" || *registerPassword == "" {
		fmt.Println("Error: Username and password must be provided for registration")
		os.Exit(1)
	}

	loadedCredentials, err := stores.LoadCredentials(encryptionKey, credential_file_path)
	if err != nil {
		logger.Errorf("Error loading credentials: %v", err)
		os.Exit(1)
	}

	err = loadedCredentials.AddUser(*registerUsername, *registerPassword, "admin")
	if err != nil {
		logger.Errorf("Error adding user: %v", err)
		os.Exit(1)
	}

	err = stores.SaveCredentials(loadedCredentials, encryptionKey, credential_file_path)
	if err != nil {
		logger.Errorf("Error saving credentials: %v", err)
		os.Exit(1)
	}

	logger.Infof("User %s registered successfully", *registerUsername)
	os.Exit(0)
}
