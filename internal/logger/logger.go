package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var (
	currentLevel = LevelInfo
	output       io.Writer = os.Stdout
	errorOutput  io.Writer = os.Stderr
)

func SetLevel(level Level) {
	currentLevel = level
}

func SetOutput(w io.Writer) {
	output = w
}

func Debug(format string, args ...interface{}) {
	if currentLevel <= LevelDebug {
		log.New(output, "[DEBUG] ", log.Ltime).Printf(format, args...)
	}
}

func Info(format string, args ...interface{}) {
	if currentLevel <= LevelInfo {
		log.New(output, "[INFO] ", log.Ltime).Printf(format, args...)
	}
}

func Warn(format string, args ...interface{}) {
	if currentLevel <= LevelWarn {
		log.New(errorOutput, "[WARN] ", log.Ltime).Printf(format, args...)
	}
}

func Error(format string, args ...interface{}) {
	if currentLevel <= LevelError {
		log.New(errorOutput, "[ERROR] ", log.Ltime).Printf(format, args...)
	}
}

func Fatal(format string, args ...interface{}) {
	log.New(errorOutput, "[FATAL] ", log.Ltime).Printf(format, args...)
	os.Exit(1)
}

// Progress muestra progreso sin timestamp
func Progress(format string, args ...interface{}) {
	fmt.Fprintf(output, format+"\n", args...)
}
