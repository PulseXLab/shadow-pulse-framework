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
	errorLogFile *os.File
	errorLogger  *log.Logger
	// Regex to strip ANSI color codes
	ansiRegex = regexp.MustCompile("(\\x1b\\[[0-9;]*m)")
)

// InitErrorLogger initializes the file logger for errors.
func InitErrorLogger(logDir string) {
	var err error
	// Log errors to error.log in the specified logDir
	logFilePath := filepath.Join(logDir, "error.log")
	errorLogFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// If we can't even open the error log, print to stderr as a last resort.
		fmt.Fprintf(os.Stderr, "Failed to open error log file: %v\n", err)
		return
	}
	errorLogger = log.New(errorLogFile, "", log.LstdFlags)
}

// CloseErrorLogger closes the error log file.
func CloseErrorLogger() {
	if errorLogFile != nil {
		errorLogFile.Close()
	}
}

func formatTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// stripColor removes ANSI color codes from a string.
func stripColor(str string) string {
	return ansiRegex.ReplaceAllString(str, "")
}

func logErrorToFile(message string) {
	if errorLogger != nil {
		errorLogger.Println(stripColor(message))
	}
}

// PrintInfo prints a standard informational message to the console.
func PrintInfo(message string) {
	formatted := fmt.Sprintf("%s[*] %s - %s%s", ColorBlue, formatTime(), message, ColorReset)
	fmt.Println(formatted)
}

// PrintGood prints a success message to the console.
func PrintGood(message string) {
	formatted := fmt.Sprintf("%s[+] %s - %s%s", ColorGreen, formatTime(), message, ColorReset)
	fmt.Println(formatted)
}

// PrintError logs an error message to error.log.
func PrintError(message string) {
	formatted := fmt.Sprintf("[!] %s - %s", formatTime(), message)
	logErrorToFile(formatted)
}

// PrintDebug logs a debug message to error.log.
func PrintDebug(message string) {
	formatted := fmt.Sprintf("[DEBUG] %s - %s", formatTime(), message)
	logErrorToFile(formatted)
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
