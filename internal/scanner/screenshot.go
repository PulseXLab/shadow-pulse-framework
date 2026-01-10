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

// TakeScreenshots runs cutycapt to capture screenshots of live web services.
func TakeScreenshots(outputDir string, useTor bool, useStealth bool) {
	utils.PrintInfo("Taking screenshots with CutyCapt...")

	liveHostsFile := filepath.Join(outputDir, "live_web_hosts.txt")
	file, err := os.Open(liveHostsFile)
	if err != nil {
		utils.PrintInfo("live_web_hosts.txt not found, skipping screenshots.")
		return
	}
	defer file.Close()

	screenshotDir := filepath.Join(outputDir, "screenshots")
	if err := os.MkdirAll(screenshotDir, os.ModePerm); err != nil {
		utils.PrintError("Failed to create screenshot directory: " + err.Error())
		return
	}

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

	for _, url := range urls {
		if useTor && useStealth {
			utils.PrintInfo(fmt.Sprintf("Requesting new Tor IP for stealth screenshot of %s", url))
			tor.RenewTorIP(outputDir)
		}

		outputFile := filepath.Join(screenshotDir, utils.SanitizeURL(url))
		// CutyCapt is simpler and does not require as many headless-related flags
		// Use xvfb-run to ensure it runs in a headless environment, and use an absolute path to avoid PATH issues.
		cmd := fmt.Sprintf(`/usr/bin/xvfb-run cutycapt --url="%s" --out="%s"`, url, outputFile)
		
		runner.RunCommand(cmd, useTor)
	}

	utils.PrintGood(fmt.Sprintf("Screenshot process complete. Reports are in %s", screenshotDir))
}
