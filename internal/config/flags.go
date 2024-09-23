package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/yashs662/SynchroDB/internal/logger"
)

type Config struct {
	Debug bool
	ID    string
	Peers []string
	Port  string
}

func ParseFlags() Config {
	debug := flag.Bool("debug", false, "enable debug mode with detailed logging")
	id := flag.String("id", "node1", "Node ID")
	peers := flag.String("peers", "", "Comma-separated list of peer addresses")
	port := flag.String("port", "8001", "Port to run the server on")
	flag.BoolVar(debug, "d", false, "enable debug mode with detailed logging (shorthand)")
	flag.StringVar(id, "i", "node1", "Node ID (shorthand)")
	flag.StringVar(peers, "p", "", "Comma-separated list of peer addresses (shorthand)")
	flag.StringVar(port, "P", "8001", "Port to run the server on (shorthand)")
	flag.Parse()

	var peerList []string
	if *peers != "" {
		peerList = strings.Split(*peers, ",")
	}

	config := Config{
		Debug: *debug,
		ID:    *id,
		Peers: peerList,
		Port:  *port,
	}

	validateConfig(config)

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

func validateConfig(config Config) {
	if config.ID == "" {
		fmt.Println("Error: Node ID cannot be empty")
		os.Exit(1)
	}

	if config.Port == "" {
		fmt.Println("Error: Port cannot be empty")
		os.Exit(1)
	}

	portNum, err := strconv.Atoi(config.Port)
	if err != nil || portNum < 1 || portNum > 65535 {
		fmt.Println("Error: Invalid port number. Must be between 1 and 65535")
		os.Exit(1)
	}

	if len(config.Peers) > 0 {
		logger.Infof("Peers: %v", config.Peers)
		logger.Infof("Number of peers: %d", len(config.Peers))
		for _, peer := range config.Peers {
			if peer == "" {
				fmt.Println("Error: Peer address cannot be empty")
				os.Exit(1)
			}
		}
	}
}
