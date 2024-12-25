package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/yashs662/SynchroDB/internal/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
)

var (
	consoleInfoLogger  *log.Logger
	consoleWarnLogger  *log.Logger
	consoleErrorLogger *log.Logger
	consoleFatalLogger *log.Logger
	consoleDebugLogger *log.Logger

	fileInfoLogger  *log.Logger
	fileWarnLogger  *log.Logger
	fileErrorLogger *log.Logger
	fileFatalLogger *log.Logger
	fileDebugLogger *log.Logger

	debugMode bool
	logQueue  chan logEntry
	wg        sync.WaitGroup
)

type logEntry struct {
	level   string
	message string
}

func Init(cfg *config.Config) {
	debugMode = cfg.Log.Debug
	initializeLoggers(cfg.Log.File)
	logQueue = make(chan logEntry, 1000)
	wg.Add(1)
	go processLogQueue()
	if debugMode {
		configJSON, _ := json.MarshalIndent(cfg, "", "  ")
		Debugf("Loaded configuration: %s", configJSON)
	}
}

func initializeLoggers(logFile string) {
	logFlags := log.Ldate | log.Ltime
	if debugMode {
		logFlags |= log.Lmicroseconds
	}

	// File logger: Plain text, no colors
	fileOutput := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
		Compress:   true,
	}
	fileInfoLogger = log.New(fileOutput, "INFO:  ", logFlags)
	fileWarnLogger = log.New(fileOutput, "WARN:  ", logFlags)
	fileErrorLogger = log.New(fileOutput, "ERROR: ", logFlags)
	fileFatalLogger = log.New(fileOutput, "FATAL: ", logFlags)
	fileDebugLogger = log.New(fileOutput, "DEBUG: ", logFlags)

	// Console logger: Colored output
	consoleInfoLogger = log.New(os.Stdout, fmt.Sprintf("%sINFO:  %s", Green, Reset), logFlags)
	consoleWarnLogger = log.New(os.Stdout, fmt.Sprintf("%sWARN:  %s", Yellow, Reset), logFlags)
	consoleErrorLogger = log.New(os.Stderr, fmt.Sprintf("%sERROR: %s", Red, Reset), logFlags)
	consoleFatalLogger = log.New(os.Stderr, fmt.Sprintf("%sFATAL: %s", Red, Reset), logFlags)
	consoleDebugLogger = log.New(os.Stdout, fmt.Sprintf("%sDEBUG: %s", Blue, Reset), logFlags)
}

func processLogQueue() {
	defer wg.Done()
	for entry := range logQueue {
		switch entry.level {
		case "INFO":
			consoleInfoLogger.Println(entry.message)
			fileInfoLogger.Println(entry.message)
		case "WARN":
			consoleWarnLogger.Println(entry.message)
			fileWarnLogger.Println(entry.message)
		case "ERROR":
			consoleErrorLogger.Println(entry.message)
			fileErrorLogger.Println(entry.message)
		case "FATAL":
			consoleFatalLogger.Fatalln(entry.message)
			fileFatalLogger.Fatalln(entry.message)
		case "DEBUG":
			if debugMode {
				consoleDebugLogger.Println(entry.message)
				fileDebugLogger.Println(entry.message)
			}
		}
	}
}

func Info(message string) {
	logQueue <- logEntry{level: "INFO", message: message}
}

func Warn(message string) {
	logQueue <- logEntry{level: "WARN", message: message}
}

func Error(message string) {
	logQueue <- logEntry{level: "ERROR", message: message}
}

func Fatal(message string) {
	logQueue <- logEntry{level: "FATAL", message: message}
	processLogQueue()
}

func Debug(message string) {
	if debugMode {
		logQueue <- logEntry{level: "DEBUG", message: message}
	}
}

func Infof(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	logQueue <- logEntry{level: "INFO", message: msg}
}

func Warnf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	logQueue <- logEntry{level: "WARN", message: msg}
}

func Errorf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	logQueue <- logEntry{level: "ERROR", message: msg}
}

func Fatalf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	logQueue <- logEntry{level: "FATAL", message: msg}
	processLogQueue()
}

func Debugf(format string, v ...interface{}) {
	if debugMode {
		msg := fmt.Sprintf(format, v...)
		logQueue <- logEntry{level: "DEBUG", message: msg}
	}
}

func StructuredInfo(fields map[string]interface{}) {
	jsonLog, _ := json.Marshal(fields)
	Info(string(jsonLog))
}

func InfoWithContext(ctx context.Context, message string) {
	requestID := ctx.Value("requestID")
	if requestID != nil {
		Infof("[RequestID: %v] %s", requestID, message)
	} else {
		Info(message)
	}
}

func SetDebugMode(debug bool) {
	debugMode = debug
}

func Close() {
	close(logQueue)
	wg.Wait()
}
