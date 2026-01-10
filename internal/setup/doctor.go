package setup

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/shadow-pulse/shadow-pulse-framework/internal/runner"
	"github.com/shadow-pulse/shadow-pulse-framework/internal/scanner"
	"github.com/shadow-pulse/shadow-pulse-framework/internal/utils"
)

// OSInfo holds detected information about the underlying operating system.
type OSInfo struct {
	Family         string // e.g., "debian", "rhel"
	PackageManager string // e.g., "apt-get", "dnf"
}

// Check represents the result of a single diagnostic check.
type Check struct {
	Description string
	Success     bool
	Message     string // Message to show on failure
	Resolution  string // Command or instruction to fix the issue
}

// RunDoctor performs all system checks and prints a diagnostic report.
func RunDoctor(fix bool) {
	utils.PrintInfo("Starting Shadow-Pulse System Diagnostic...")

	osInfo := detectOS()
	if osInfo.Family == "unknown" {
		utils.PrintError("Could not determine OS family. Cannot provide specific remediation steps.")
	} else {
		utils.PrintGood(fmt.Sprintf("Detected OS Family: %s", osInfo.Family))
	}

	var checks []Check
	remediationSteps := make(map[string]string) // Use map to avoid duplicate resolutions

	// Perform all checks
	checks = append(checks, checkTools(osInfo)...)
	checks = append(checks, checkSecLists(osInfo))
	checks = append(checks, checkTor(osInfo)...)

	// --- Print Report ---
	utils.PrintInfo("\nDiagnostic Report:")
	allClear := true

	for _, check := range checks {
		if check.Success {
			utils.PrintGood(fmt.Sprintf("[✓] %s", check.Description))
		} else {
			utils.PrintError(fmt.Sprintf("[✗] %s", check.Description))
			if check.Message != "" {
				utils.PrintError(fmt.Sprintf("  └─ Problem: %s", check.Message))
			}
			if check.Resolution != "" {
				remediationSteps[check.Description] = check.Resolution
			}
			allClear = false
		}
	}

	fmt.Println() // Spacer

	if allClear {
		utils.PrintGood("🎉 System is ready! All checks passed.")
	} else {
		if len(remediationSteps) > 0 {
			utils.PrintInfo("Found one or more issues. Please follow the steps below to resolve them.")
			fmt.Println("==================================================")
			fmt.Println("              💊 PRESCRIPTION 💊")
			fmt.Println("==================================================")
			i := 1
			for desc, step := range remediationSteps {
				fmt.Printf("%d. To fix '%s':\n", i, desc)
				fmt.Printf("   %s\n\n", step)
				i++
			}
			fmt.Println("==================================================")
			
			if fix {
				utils.PrintInfo("Attempting to automatically install missing dependencies...")
				for desc, step := range remediationSteps {
					utils.PrintInfo(fmt.Sprintf("  -> Installing for '%s': %s", desc, step))
					// Installation commands should not go through Tor
					err := runner.RunCommand(step, false) 
					if err != nil {
						utils.PrintError(fmt.Sprintf("Failed to automatically install '%s'. Please try manually: %s", desc, step))
					} else {
						utils.PrintGood(fmt.Sprintf("  -> Successfully attempted installation for '%s'.", desc))
					}
				}
				utils.PrintInfo("Automatic installation attempts finished. Please run 'shadow-pulse doctor' again to verify.")
			} else {
				utils.PrintInfo("After running the commands, please run 'shadow-pulse doctor' again to verify.")
			}

		} else {
			utils.PrintError("One or more checks failed, but no automatic resolution is available.")
		}
	}
}

// checkTools verifies all required external tools are installed.
func checkTools(osInfo OSInfo) []Check {
	var checks []Check
	for _, tool := range scanner.RequiredTools {
		check := Check{Description: fmt.Sprintf("Tool '%s' is installed", tool)}
		if runner.LookPath(tool) {
			check.Success = true
		} else {
			check.Success = false
			check.Message = "Tool not found in system PATH."
			if cmd, ok := scanner.ToolInstallCommands[tool]; ok {
				// Adjust command for non-debian systems if needed
				if osInfo.Family == "rhel" {
					// Example adjustment, can be expanded
					cmd = strings.Replace(cmd, "apt-get update && sudo apt-get install -y", "dnf install -y", 1)
				}
				check.Resolution = cmd
			}
		}
		checks = append(checks, check)
	}
	return checks
}

// checkSecLists verifies the presence of the Seclists wordlist.
func checkSecLists(osInfo OSInfo) Check {
	check := Check{Description: "Seclists wordlist for gobuster"}
	wordlistPath := "/usr/share/seclists/Discovery/DNS/subdomains-top1million-5000.txt"
	if _, err := os.Stat(wordlistPath); err == nil {
		check.Success = true
	} else {
		check.Success = false
		check.Message = "Wordlist not found at " + wordlistPath
		if osInfo.Family == "debian" {
			check.Resolution = "sudo apt-get update && sudo apt-get install -y seclists"
		}
	}
	return check
}

// checkTor performs a deep diagnostic on the Tor service and configuration.
func checkTor(osInfo OSInfo) []Check {
	var checks []Check

	// 1. Check if service is active
	cmd := exec.Command("systemctl", "is-active", "tor")
	err := cmd.Run()
	checkSvc := Check{
		Description: "Tor service is active and running",
		Success:     err == nil,
		Message:     "Tor service is not running or has failed.",
		Resolution:  "sudo systemctl restart tor",
	}
	checks = append(checks, checkSvc)
	// If service isn't running, further checks are pointless
	if !checkSvc.Success {
		return checks
	}

	// 2. Check ports
	checkSocks := Check{
		Description: "Tor SOCKS port (9050) is reachable",
		Success:     isPortOpen("127.0.0.1:9050"),
		Message:     "Cannot connect to SOCKS port. Check torrc for 'SocksPort' and service logs with 'sudo journalctl -u tor'.",
	}
	checkControl := Check{
		Description: "Tor ControlPort (9051) is reachable",
		Success:     isPortOpen("127.0.0.1:9051"),
		Message:     "Cannot connect to ControlPort. Check torrc for 'ControlPort 9051' and service logs.",
	}
	checks = append(checks, checkSocks, checkControl)

	return checks
}

// isPortOpen checks if a TCP port is open on a given address.
func isPortOpen(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}


// detectOS reads /etc/os-release to determine the OS family and package manager.
func detectOS() OSInfo {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return OSInfo{Family: "unknown"}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	vars := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := strings.Trim(parts[1], `"`)
			vars[key] = value
		}
	}

	if id, ok := vars["ID"]; ok {
		switch id {
		case "ubuntu", "debian", "kali", "raspbian":
			return OSInfo{Family: "debian", PackageManager: "apt-get"}
		case "centos", "rhel", "fedora":
			return OSInfo{Family: "rhel", PackageManager: "dnf"}
		}
	}
	
	if idLike, ok := vars["ID_LIKE"]; ok {
		if strings.Contains(idLike, "debian") {
			return OSInfo{Family: "debian", PackageManager: "apt-get"}
		}
		if strings.Contains(idLike, "rhel") || strings.Contains(idLike, "fedora") {
			return OSInfo{Family: "rhel", PackageManager: "dnf"}
		}
	}

	return OSInfo{Family: "unknown"}
}
