package database

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type AOFWriter struct {
	file *os.File
	mu   sync.Mutex
}

func NewAOFWriter(filepath string) (*AOFWriter, error) {
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &AOFWriter{file: file}, nil
}

func (aof *AOFWriter) Write(command string) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()
	timestamp := time.Now().Unix()
	_, err := aof.file.WriteString(fmt.Sprintf("%d %s\n", timestamp, command))
	return err
}

func (aof *AOFWriter) Close() error {
	return aof.file.Close()
}
