package scanner

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/shadow-pulse/shadow-pulse-framework/internal/runner"
	"github.com/shadow-pulse/shadow-pulse-framework/internal/utils"
)

// RunVulnerabilityScans reads hosts from the live_web_hosts.txt file and runs scanners.
func RunVulnerabilityScans(outputDir string, useTor bool) {
	hostsFile := filepath.Join(outputDir, "live_web_hosts.txt")
	file, err := os.Open(hostsFile)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to open live hosts file for vuln scanning: %s", err))
		return
	}
	defer file.Close()

	var liveServers []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			liveServers = append(liveServers, line)
		}
	}

	if err := scanner.Err(); err != nil {
		utils.PrintError(fmt.Sprintf("Error reading live hosts file: %s", err))
		return
	}

	if len(liveServers) == 0 {
		utils.PrintInfo("No live web servers found in file to scan for vulnerabilities.")
		return
	}

	utils.PrintInfo("Starting web vulnerability scans on hosts from live_web_hosts.txt...")

	var wg sync.WaitGroup
	for _, serverURL := range liveServers {
		wg.Add(3) // Adding 3 for nikto, wpscan, and nuclei

		go runNikto(serverURL, outputDir, useTor, &wg)
		go runWpscan(serverURL, outputDir, useTor, &wg)
		go runNuclei(serverURL, outputDir, useTor, &wg)
	}
	wg.Wait()
	utils.PrintGood("Web vulnerability scanning completed.")
}

// runNikto runs a Nikto scan on a single URL.
func runNikto(serverURL, outputDir string, useTor bool, wg *sync.WaitGroup) {
	defer wg.Done()
	utils.PrintInfo(fmt.Sprintf("Running nikto on %s", serverURL))
	safeFilename := utils.SanitizeFilename(serverURL)
	outputFile := filepath.Join(outputDir, fmt.Sprintf("nikto_%s.txt", safeFilename))
	cmd := fmt.Sprintf("nikto -h %s -output %s", serverURL, outputFile)
	runner.RunCommand(cmd, useTor)
}

// runWpscan runs a WPScan on a single URL.
func runWpscan(serverURL, outputDir string, useTor bool, wg *sync.WaitGroup) {
	defer wg.Done()
	// Basic check to see if it's worth running wpscan.
	if !strings.Contains(strings.ToLower(serverURL), "wp-") && !strings.Contains(strings.ToLower(serverURL), "wordpress") {
		// Silently skip if it doesn't seem to be a WordPress site.
		return
	}
	utils.PrintInfo(fmt.Sprintf("Running wpscan on %s", serverURL))
	safeFilename := utils.SanitizeFilename(serverURL)
	outputFile := filepath.Join(outputDir, fmt.Sprintf("wpscan_%s.txt", safeFilename))
	// WPScan requires an API token for full results, not included here.
	cmd := fmt.Sprintf("wpscan --url %s --random-user-agent --disable-tls-checks -o %s", serverURL, outputFile)
	runner.RunCommand(cmd, useTor)
}

// runNuclei runs a Nuclei scan on a single URL.
func runNuclei(serverURL, outputDir string, useTor bool, wg *sync.WaitGroup) {
	defer wg.Done()
	utils.PrintInfo(fmt.Sprintf("Running nuclei on %s", serverURL))
	safeFilename := utils.SanitizeFilename(serverURL)
	outputFile := filepath.Join(outputDir, fmt.Sprintf("nuclei_%s.txt", safeFilename))
	cmd := fmt.Sprintf("nuclei -u %s -o %s", serverURL, outputFile)
	runner.RunCommand(cmd, useTor)
}
