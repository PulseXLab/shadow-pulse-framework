package scanner

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shadow-pulse/shadow-pulse-framework/internal/runner"
	"github.com/shadow-pulse/shadow-pulse-framework/internal/utils"
)

// FindLiveWebServers uses httpx to check for live web servers from a list of subdomains.
func FindLiveWebServers(subdomains []string, outputDir string, useTor bool) []string {
	utils.PrintInfo("Checking for live web servers with httpx...")

	if len(subdomains) == 0 {
		utils.PrintError("No subdomains provided to check for live web servers.")
		return nil
	}

	// Create a temporary file with the list of subdomains
	inputFile := filepath.Join(outputDir, "httpx_input.txt")
	file, err := os.Create(inputFile)
	if err != nil {
		utils.PrintError("Failed to create input file for httpx: " + err.Error())
		return nil
	}
	for _, sub := range subdomains {
		fmt.Fprintln(file, sub)
	}
	file.Close() // Close the file so httpx can read it

	outputFile := filepath.Join(outputDir, "live_web_hosts.txt")
	cmd := fmt.Sprintf("httpx -l %s -threads 50 -follow-redirects -silent -o %s", inputFile, outputFile)
	
	// Note: httpx has its own proxy support, but proxychains should also work.
	// For simplicity, we continue to use the global 'useTor' flag.
	runner.RunCommand(cmd, useTor)

	// Read the results from the output file
	var liveServers []string
	resultsFile, err := os.Open(outputFile)
	if err != nil {
		utils.PrintError("Failed to open httpx output file: " + err.Error())
		return nil
	}
	defer resultsFile.Close()

	scanner := bufio.NewScanner(resultsFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			liveServers = append(liveServers, line)
		}
	}

	if len(liveServers) == 0 {
		utils.PrintGood("No live web servers found by httpx.")
	} else {
		utils.PrintGood(fmt.Sprintf("Found %d live web servers.", len(liveServers)))
	}

	return liveServers
}
