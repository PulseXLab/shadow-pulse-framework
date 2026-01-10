package scanner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shadow-pulse/shadow-pulse-framework/internal/runner"
	"github.com/shadow-pulse/shadow-pulse-framework/internal/utils"
)

// TakeScreenshots runs eyewitness to capture screenshots of live web services.
func TakeScreenshots(outputDir string, useTor bool) {
	utils.PrintInfo("Taking screenshots with eyewitness...")
	
	liveHostsFile := filepath.Join(outputDir, "live_web_hosts.txt")
	if _, err := os.Stat(liveHostsFile); err != nil {
		utils.PrintInfo("live_web_hosts.txt not found (no live web servers discovered by httpx), skipping screenshots.")
		return
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
