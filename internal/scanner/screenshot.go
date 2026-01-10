package scanner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shadow-pulse/shadow-pulse-framework/internal/runner"
	"github.com/shadow-pulse/shadow-pulse-framework/internal/utils"
)

// TakeScreenshots runs eyewitness to capture screenshots of live web services.
func TakeScreenshots(liveHosts []string, outputDir string, useTor bool) {
	utils.PrintInfo("Taking screenshots with eyewitness...")

	if len(liveHosts) == 0 {
		utils.PrintInfo("No live web hosts found, skipping screenshots.")
		return
	}

	liveHostsFile := filepath.Join(outputDir, "live_web_hosts.txt")
	file, err := os.Create(liveHostsFile)
	if err != nil {
		utils.PrintError("Failed to create live_web_hosts.txt: " + err.Error())
		return
	}
	defer file.Close()

	for _, host := range liveHosts {
		fmt.Fprintln(file, host)
	}

	screenshotDir := filepath.Join(outputDir, "eyewitness_report")

	// Use absolute path for eyewitness directory to avoid potential issues
	absScreenshotDir, err := filepath.Abs(screenshotDir)
	if err != nil {
		utils.PrintError("Failed to get absolute path for screenshot directory: " + err.Error())
		// Fallback to relative path
		absScreenshotDir = screenshotDir
	}

	seleniumLogPath := filepath.Join(outputDir, "geckodriver.log")
	cmd := fmt.Sprintf("eyewitness -f %s -d %s --web --no-prompt --selenium-log-path %s", liveHostsFile, absScreenshotDir, seleniumLogPath)
	runner.RunCommand(cmd, useTor)

	utils.PrintGood(fmt.Sprintf("Eyewitness report should be available in %s", absScreenshotDir))
}
