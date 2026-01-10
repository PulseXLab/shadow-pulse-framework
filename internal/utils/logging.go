package utils

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"time"
)

const (
	ColorReset  = "\033[0m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorRed    = "\033[31m"
	ColorBlue   = "\033[34m"
)

var (
	logFile *os.File
	logger  *log.Logger
	// Regex to strip ANSI color codes
	ansiRegex = regexp.MustCompile("(\\x1b\\[[0-9;]*m)")
)

// InitLogger initializes the file logger.
func InitLogger(logFilePath string) {
	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		PrintError("Failed to open log file: " + err.Error())
		// Continue without file logging
		return
	}
	logger = log.New(logFile, "", 0) // No prefix, we handle it
	PrintGood("Application execution log will be saved to " + logFilePath)
}

// CloseLogger closes the log file.
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

func formatTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// stripColor removes ANSI color codes from a string.
func stripColor(str string) string {
	return ansiRegex.ReplaceAllString(str, "")
}

func logToFile(formattedMessage string) {
	if logger != nil {
		logger.Println(stripColor(formattedMessage))
	}
}

// PrintInfo prints a standard informational message.
func PrintInfo(message string) {
	formatted := fmt.Sprintf("%s[*] %s - %s%s", ColorBlue, formatTime(), message, ColorReset)
	fmt.Println(formatted)
	logToFile(formatted)
}

// PrintGood prints a success message.
func PrintGood(message string) {
	formatted := fmt.Sprintf("%s[+] %s - %s%s", ColorGreen, formatTime(), message, ColorReset)
	fmt.Println(formatted)
	logToFile(formatted)
}

// PrintError prints an error message.
func PrintError(message string) {
	formatted := fmt.Sprintf("%s[!] %s - %s%s", ColorRed, formatTime(), message, ColorReset)
	fmt.Println(formatted)
	logToFile(formatted)
}

// PrintDebug prints a debug message.
func PrintDebug(message string) {
	formatted := fmt.Sprintf("%s[DEBUG] %s - %s%s", ColorYellow, formatTime(), message, ColorReset)
	fmt.Println(formatted)
	logToFile(formatted)
}

// AppendToFile opens a file in append mode and writes content to it.
func AppendToFile(filePath, content string) {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to open file for appending (%s): %v", filePath, err))
		return
	}
	defer f.Close()

	if _, err := f.WriteString(content + "\n"); err != nil {
		PrintError(fmt.Sprintf("Failed to write to file (%s): %v", filePath, err))
	}
}
