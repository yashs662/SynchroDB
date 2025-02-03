package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/yashs662/SynchroDB/internal/config"
	"github.com/yashs662/SynchroDB/pkg/client"
	"github.com/yashs662/SynchroDB/pkg/protocol"
)

var (
	benchmarkPrefix = "synchrodb-benchmark"
	pingCommand     = protocol.PingCommand{}
	setCommand      = protocol.SetCommand{}
	getCommand      = protocol.GetCommand{}
	incrCommand     = protocol.IncrCommand{}
	decrCommand     = protocol.DecrCommand{}
)

var defaultCommands = []string{
	pingCommand.GetCommandInfo().Command,
	fmt.Sprintf("%s %s:test 123", setCommand.GetCommandInfo().Command, benchmarkPrefix),
	fmt.Sprintf("%s %s:test", getCommand.GetCommandInfo().Command, benchmarkPrefix),
	fmt.Sprintf("%s %s:test", incrCommand.GetCommandInfo().Command, benchmarkPrefix),
	fmt.Sprintf("%s %s:test", decrCommand.GetCommandInfo().Command, benchmarkPrefix),
}

func main() {
	address := flag.String("address", "127.0.0.1:8000", "Server address")
	configPath := flag.String("config", "config/server.yaml", "Path to the server config file")
	benchmark := flag.Bool("benchmark", false, "Benchmark the command")
	clients := flag.Int("clients", 10, "Number of concurrent clients for benchmarking")
	iterations := flag.Int("iterations", 1000, "Number of iterations per client for benchmarking")

	flag.Parse()

	cfg, err := config.LoadConfigFromPath(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	client, err := client.NewClient(*address, cfg.Server.Password, cfg.Server.AuthEnabled)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	if *benchmark {
		commands := defaultCommands
		results, successfulClients, totalCommands, duration, err := client.Benchmark(commands, *clients, *iterations)
		if err != nil {
			log.Fatalf("Benchmark failed: %v", err)
		}
		printBenchmarkResults(results, successfulClients, totalCommands, duration, *clients, *iterations)
	} else {
		interactiveMode(client)
	}
}

func interactiveMode(client *client.Client) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Entering interactive mode. Type 'exit' to quit.")
	for {
		fmt.Print("> ")
		command, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Failed to read input: %v", err)
		}
		command = strings.TrimSpace(command)
		// TODO: Add support for more client specific commands like clear, up down navigation for history etc.
		if command == "exit" {
			break
		}
		response, err := client.SendCommand(command)
		if err != nil {
			color.Red("Error: %v\n", err)
		} else {
			if strings.HasPrefix(response, "ERR") {
				color.Red("Response: %s\n", response)
			} else {
				color.Green("Response: %s\n", response)
			}
		}
	}
}

func printBenchmarkResults(results map[string]client.BenchmarkResult, successfulClients, totalCommands int, duration time.Duration, clients, iterations int) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Command", "Min (ms)", "Max (ms)", "Avg (ms)", "P99 (ms)", "Throughput (ops/sec)"})

	for command, result := range results {
		throughput := float64(totalCommands) / duration.Seconds()
		table.Append([]string{
			command,
			fmt.Sprintf("%.3f", result.Min.Seconds()*1000),
			fmt.Sprintf("%.3f", result.Max.Seconds()*1000),
			fmt.Sprintf("%.3f", result.Avg.Seconds()*1000),
			fmt.Sprintf("%.3f", result.P99.Seconds()*1000),
			fmt.Sprintf("%.3f", throughput),
		})
	}

	table.Render()
	fmt.Printf("Successful clients: %d/%d\n", successfulClients, clients)
	fmt.Printf("Iterations per client: %d\n", iterations)
	fmt.Printf("Total commands executed: %d\n", totalCommands)
	fmt.Printf("Total duration: %.2f seconds\n", duration.Seconds())
}
