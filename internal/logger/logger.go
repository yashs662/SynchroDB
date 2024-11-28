package logger

import (
	"fmt"
	"io"
	"log"
	"os"

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
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
	debugMode   bool
)

func Init(debug bool) {
	debugMode = debug
	initializeLoggers()
}

func initializeLoggers() {
	logFlags := log.Ldate | log.Ltime | log.Lmicroseconds
	if debugMode {
		logFlags |= log.Lshortfile // Add file name and line number
	}

	logOutput := &lumberjack.Logger{
		Filename:   "synchrodb.log",
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
		Compress:   true,
	}

	multiWriter := io.MultiWriter(os.Stdout, logOutput)

	infoLogger = log.New(multiWriter, fmt.Sprintf("%sINFO:  %s", Green, Reset), logFlags)
	warnLogger = log.New(multiWriter, fmt.Sprintf("%sWARN:  %s", Yellow, Reset), logFlags)
	errorLogger = log.New(multiWriter, fmt.Sprintf("%sERROR: %s", Red, Reset), logFlags)
	debugLogger = log.New(multiWriter, fmt.Sprintf("%sDEBUG: %s", Blue, Reset), logFlags)

	// Loggers without color for file output
	fileInfoLogger := log.New(logOutput, "INFO:  ", logFlags)
	fileWarnLogger := log.New(logOutput, "WARN:  ", logFlags)
	fileErrorLogger := log.New(logOutput, "ERROR: ", logFlags)
	fileDebugLogger := log.New(logOutput, "DEBUG: ", logFlags)

	// Set output for each logger
	infoLogger.SetOutput(io.MultiWriter(os.Stdout, fileInfoLogger.Writer()))
	warnLogger.SetOutput(io.MultiWriter(os.Stdout, fileWarnLogger.Writer()))
	errorLogger.SetOutput(io.MultiWriter(os.Stderr, fileErrorLogger.Writer()))
	debugLogger.SetOutput(io.MultiWriter(os.Stdout, fileDebugLogger.Writer()))
}

func Info(message string) {
	infoLogger.Println(message)
}

func Warn(message string) {
	warnLogger.Println(message)
}

func Error(message string) {
	errorLogger.Println(message)
}

func Debug(message string) {
	if debugMode {
		debugLogger.Println(message)
	}
}

func Infof(format string, v ...interface{}) {
	infoLogger.Printf(format, v...)
}

func Warnf(format string, v ...interface{}) {
	warnLogger.Printf(format, v...)
}

func Errorf(format string, v ...interface{}) {
	errorLogger.Printf(format, v...)
}

func Debugf(format string, v ...interface{}) {
	if debugMode {
		debugLogger.Printf(format, v...)
	}
}

func SetDebugMode(debug bool) {
	debugMode = debug
	initializeLoggers()
}
