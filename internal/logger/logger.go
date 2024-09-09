package logger

import (
	"fmt"
	"log"
	"os"
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
	logFlags := log.Ldate | log.Ltime
	if debugMode {
		logFlags |= log.Lshortfile // Add file name and line number
	}

	infoLogger = log.New(os.Stdout, fmt.Sprintf("%sINFO:  %s", Green, Reset), logFlags)
	warnLogger = log.New(os.Stdout, fmt.Sprintf("%sWARN:  %s", Yellow, Reset), logFlags)
	errorLogger = log.New(os.Stderr, fmt.Sprintf("%sERROR: %s", Red, Reset), logFlags)
	debugLogger = log.New(os.Stdout, fmt.Sprintf("%sDEBUG: %s", Blue, Reset), logFlags)
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
