package tor

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shadow-pulse/shadow-pulse-framework/internal/utils"
	"golang.org/x/net/proxy"
)

const torAuthCookiePath = "/var/run/tor/control.authcookie" // Common path for Tor control cookie

// GetTorIP gets the current external IP address via the Tor proxy.
func GetTorIP() string {
	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, proxy.Direct)
	if err != nil {
		utils.PrintError("Failed to create SOCKS5 dialer.")
		return "unknown"
	}

	httpTransport := &http.Transport{
		Dial: dialer.Dial,
	}
	httpClient := &http.Client{
		Transport: httpTransport,
		Timeout:   15 * time.Second,
	}

	resp, err := httpClient.Get("https://httpbin.org/ip")
	if err != nil {
		utils.PrintError("Failed to get current Tor IP: " + err.Error())
		return "unknown"
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.PrintError("Failed to read IP response body.")
		return "unknown"
	}

	// Very basic parsing, assumes format {"origin": "x.x.x.x"}
	ip := string(body)
	// Add robustness for different json responses
	ip = strings.Replace(ip, `{"origin": "`, "", 1)
	ip = strings.Split(ip, "\"")[0]
	
	return ip
}

// RenewTorIP signals Tor to renew its IP address and logs the new IP.
func RenewTorIP(outputDir string) {
	utils.PrintInfo("Requesting new Tor IP address...")
	conn, err := net.Dial("tcp", "127.0.0.1:9051")
	if err != nil {
		utils.PrintError("Cannot connect to Tor ControlPort. Is it enabled and listening on 127.0.0.1:9051?")
		return
	}
	defer conn.Close()

	// Read authentication cookie
	cookie, err := os.ReadFile(torAuthCookiePath)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to read Tor control cookie from %s: %v", torAuthCookiePath, err))
		return
	}
	cookieHex := hex.EncodeToString(cookie)

	// Tor controller authentication
	fmt.Fprintf(conn, "AUTHENTICATE %s\r\n", cookieHex)
	status, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		utils.PrintError("Failed to read Tor ControlPort authentication response.")
		return
	}
	if !strings.Contains(status, "250") {
		utils.PrintError("Tor ControlPort authentication failed: " + status)
		return
	}

	// Send NEWNYM signal
	fmt.Fprintf(conn, "SIGNAL NEWNYM\r\n")
	status, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		utils.PrintError("Failed to read Tor ControlPort NEWNYM response.")
		return
	}
	if !strings.Contains(status, "250") {
		utils.PrintError("Failed to send NEWNYM signal: " + status)
		return
	}

	utils.PrintGood("Tor IP address has been renewed.")
	
	// Give Tor a moment to establish a new circuit
	time.Sleep(5 * time.Second) 
	
	newIP := GetTorIP()
	if newIP != "unknown" {
		logLine := fmt.Sprintf("%s - %s", time.Now().Format(time.RFC3339), newIP)
		logFilePath := filepath.Join(outputDir, "scan_tor_ip.txt")
		utils.AppendToFile(logFilePath, logLine)
		utils.PrintGood("New Tor IP appears to be: " + newIP)
	}
}

// CheckTorPrerequisites verifies that the required Tor ports are accessible.
func CheckTorPrerequisites() bool {
	utils.PrintInfo("Checking Tor prerequisites...")

	// Check SOCKS Proxy Port
	connSocks, err := net.DialTimeout("tcp", "127.0.0.1:9050", 2*time.Second)
	if err != nil {
		utils.PrintError("Tor SOCKS proxy (127.0.0.1:9050) is not reachable.")
		utils.PrintError("Please ensure the Tor service is running correctly.")
		return false
	}
	defer connSocks.Close()
	utils.PrintGood("Tor SOCKS proxy (9050) is reachable.")

	// Check Control Port
	connControl, err := net.DialTimeout("tcp", "127.0.0.1:9051", 2*time.Second)
	if err != nil {
		utils.PrintError("Tor ControlPort (127.0.0.1:9051) is not reachable.")
		utils.PrintError("Please ensure ControlPort is enabled in your torrc configuration.")
		return false
	}
	defer connControl.Close()
	utils.PrintGood("Tor ControlPort (9051) is reachable.")

	return true
}
