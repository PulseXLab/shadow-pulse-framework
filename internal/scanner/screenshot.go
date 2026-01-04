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
	
	liveHostsFile := filepath.Join(outputDir, "live_web_hosts.txt")
	screenshotDir := filepath.Join(outputDir, "eyewitness_report")

	if _, err := os.Stat(liveHostsFile); err != nil {
		utils.PrintInfo("live_web_hosts.txt not found, skipping screenshots.")
		return
	}

	cmd := fmt.Sprintf("eyewitness -f %s -d %s --web --no-prompt", liveHostsFile, screenshotDir)
	runner.RunCommand(cmd, useTor)
	
	utils.PrintGood(fmt.Sprintf("Eyewitness report should be available in %s", screenshotDir))
}
