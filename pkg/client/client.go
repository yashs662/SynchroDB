package client

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
)

type Client struct {
	conn        net.Conn
	reader      *bufio.Reader
	password    string
	authEnabled bool
}

type BenchmarkResult struct {
	Min time.Duration
	Max time.Duration
	Avg time.Duration
	P99 time.Duration
}

func NewClient(address, password string, authEnabled bool) (*Client, error) {
	conn, err := tls.Dial("tcp", address, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	client := &Client{
		conn:        conn,
		reader:      bufio.NewReader(conn),
		password:    password,
		authEnabled: authEnabled,
	}

	if authEnabled {
		if err := client.authenticate(); err != nil {
			return nil, err
		}
	}

	return client, nil
}

func (c *Client) authenticate() error {
	response, err := c.SendCommand(fmt.Sprintf("AUTH %s", c.password))
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}
	if response != "OK" {
		return fmt.Errorf("authentication failed: %s", response)
	}
	return nil
}

func (c *Client) SendCommand(command string) (string, error) {
	_, err := c.conn.Write([]byte(command + "\n"))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	response, err := c.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return strings.TrimSpace(response), nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Benchmark(commands []string, clients, iterations int) (map[string]BenchmarkResult, int, int, time.Duration, error) {
	var wg sync.WaitGroup
	results := make(map[string][]time.Duration)
	mu := sync.Mutex{}
	totalCommands := 0
	successfulClients := 0

	start := time.Now()

	for i := 0; i < clients; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client, err := NewClient(c.conn.RemoteAddr().String(), c.password, c.authEnabled)
			if err != nil {
				return
			}
			defer client.Close()

			mu.Lock()
			successfulClients++
			mu.Unlock()

			for j := 0; j < iterations; j++ {
				for _, command := range commands {
					mainCommand := strings.Split(command, " ")[0]
					start := time.Now()
					_, err := client.SendCommand(command)
					duration := time.Since(start)
					if err != nil {
						return
					}
					mu.Lock()
					results[mainCommand] = append(results[mainCommand], duration)
					totalCommands++
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	benchmarkResults := make(map[string]BenchmarkResult)
	for command, durations := range results {
		sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
		min := durations[0]
		max := durations[len(durations)-1]
		sum := time.Duration(0)
		for _, d := range durations {
			sum += d
		}
		avg := sum / time.Duration(len(durations))
		p99 := durations[int(float64(len(durations))*0.99)]

		benchmarkResults[command] = BenchmarkResult{
			Min: min,
			Max: max,
			Avg: avg,
			P99: p99,
		}
	}

	return benchmarkResults, successfulClients, totalCommands, elapsed, nil
}
