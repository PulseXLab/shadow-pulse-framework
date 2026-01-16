package scanner

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shadow-pulse/shadow-pulse-framework/internal/runner"
	"github.com/shadow-pulse/shadow-pulse-framework/internal/tor"
	"github.com/shadow-pulse/shadow-pulse-framework/internal/utils"
)

// TakeScreenshots runs eyewitness to capture screenshots of live web services.
func TakeScreenshots(outputDir string, useTor bool, useStealth bool) {
	utils.PrintInfo("Taking screenshots with Eyewitness...")

	liveHostsFile := filepath.Join(outputDir, "live_web_hosts.txt")
	if _, err := os.Stat(liveHostsFile); err != nil {
		utils.PrintInfo("live_web_hosts.txt not found, skipping screenshots.")
		return
	}

	screenshotDir := filepath.Join(outputDir, "eyewitness_report")
	
	// If not in stealth mode, run eyewitness once on the whole file for efficiency.
	if !useStealth {
		var cmd string
		if useTor {
			utils.PrintInfo("Routing eyewitness through Tor using its native SOCKS5 proxy support.")
			cmd = fmt.Sprintf("eyewitness -f %s -d %s --web --timeout 60 --no-prompt --proxy-type 'socks5' --proxy-ip '127.0.0.1' --proxy-port '9050'", liveHostsFile, screenshotDir)
		} else {
			cmd = fmt.Sprintf("eyewitness -f %s -d %s --web --timeout 60 --no-prompt", liveHostsFile, screenshotDir)
		}
		// We pass useTor=false because we are NOT using proxychains. Eyewitness handles the proxy.
		if err := runner.RunCommand(cmd, false); err != nil {
			utils.PrintError(fmt.Sprintf("Eyewitness execution failed: %v", err))
			// Decide if we should return or continue
			// For now, let's just log and continue with the next URL in stealth mode
		}
		utils.PrintGood(fmt.Sprintf("Screenshot process complete. Reports are in %s", screenshotDir))
		return
	}

	// --- Stealth Mode: Rotate IP for each URL ---
	utils.PrintInfo("Stealth mode enabled for screenshots: renewing Tor IP for each URL.")
	file, err := os.Open(liveHostsFile)
	if err != nil {
		utils.PrintError("Failed to open live_web_hosts.txt to read URLs: " + err.Error())
		return
	}
	defer file.Close()
	
	urls := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	if len(urls) == 0 {
		utils.PrintInfo("No URLs to screenshot.")
		return
	}

	utils.PrintInfo(fmt.Sprintf("Found %d URLs to screenshot.", len(urls)))
	singleURLFile := filepath.Join(outputDir, "single_url_for_eyewitness.txt")

	for _, url := range urls {
		if useTor && useStealth { // This condition is always true in this block, but good for clarity
			utils.PrintInfo(fmt.Sprintf("Requesting new Tor IP for stealth screenshot of %s", url))
			tor.RenewTorIP(outputDir)
		}

		if err := os.WriteFile(singleURLFile, []byte(url), 0644); err != nil {
			utils.PrintError("Failed to create single URL file for eyewitness: " + err.Error())
			continue
		}
		
		// Always use the proxy flags in stealth mode with Tor
		cmd := fmt.Sprintf("eyewitness -f %s -d %s --web --timeout 60 --no-prompt --proxy-type 'socks5' --proxy-ip '127.0.0.1' --proxy-port '9050'", singleURLFile, screenshotDir)
		
		// We pass useTor=false because we are NOT using proxychains.
		if err := runner.RunCommand(cmd, false); err != nil {
			utils.PrintError(fmt.Sprintf("Eyewitness execution failed: %v", err))
			// Decide if we should return or continue
			// For now, let's just log and continue with the next URL in stealth mode
		} 
	}
	// Clean up the temporary file
	os.Remove(singleURLFile)

	utils.PrintGood(fmt.Sprintf("Screenshot process complete. Reports are in %s", screenshotDir))
}
