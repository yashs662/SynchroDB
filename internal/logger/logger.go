package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

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
)

func Init(cfg *config.Config) {
	debugMode = cfg.Log.Debug
	initializeLoggers(cfg.Log.File)
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

func Info(message string) {
	consoleInfoLogger.Println(message)
	fileInfoLogger.Println(message)
}

func Warn(message string) {
	consoleWarnLogger.Println(message)
	fileWarnLogger.Println(message)
}

func Error(message string) {
	consoleErrorLogger.Println(message)
	fileErrorLogger.Println(message)
}

func Fatal(message string) {
	consoleFatalLogger.Fatalln(message)
	fileFatalLogger.Fatalln(message)
}

func Debug(message string) {
	if debugMode {
		consoleDebugLogger.Println(message)
		fileDebugLogger.Println(message)
	}
}

func Infof(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	consoleInfoLogger.Println(msg)
	fileInfoLogger.Println(msg)
}

func Warnf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	consoleWarnLogger.Println(msg)
	fileWarnLogger.Println(msg)
}

func Errorf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	consoleErrorLogger.Println(msg)
	fileErrorLogger.Println(msg)
}

func Fatalf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	consoleFatalLogger.Fatalln(msg)
	fileFatalLogger.Fatalln(msg)
}

func Debugf(format string, v ...interface{}) {
	if debugMode {
		msg := fmt.Sprintf(format, v...)
		consoleDebugLogger.Println(msg)
		fileDebugLogger.Println(msg)
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
