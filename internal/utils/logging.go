package utils

import (
	"fmt"
	"time"
)

const (
	ColorReset  = "\033[0m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorRed    = "\033[31m"
	ColorBlue   = "\033[34m"
)

func formatTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// PrintInfo prints a standard informational message.
func PrintInfo(message string) {
	fmt.Printf("%s[*] %s - %s%s\n", ColorBlue, formatTime(), message, ColorReset)
}

// PrintGood prints a success message.
func PrintGood(message string) {
	fmt.Printf("%s[+] %s - %s%s\n", ColorGreen, formatTime(), message, ColorReset)
}

// PrintError prints an error message.
func PrintError(message string) {
	fmt.Printf("%s[!] %s - %s%s\n", ColorRed, formatTime(), message, ColorReset)
}
