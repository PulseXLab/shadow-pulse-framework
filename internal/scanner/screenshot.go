package scanner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shadow-pulse/shadow-pulse-framework/internal/runner"
	"github.com/shadow-pulse/shadow-pulse-framework/internal/utils"
)

// TakeScreenshots runs goneshot to capture screenshots of live web services.
func TakeScreenshots(outputDir string, useTor bool) {
	utils.PrintInfo("Taking screenshots with goneshot...")
	
	liveHostsFile := filepath.Join(outputDir, "live_web_hosts.txt")
	if _, err := os.Stat(liveHostsFile); err != nil {
		utils.PrintInfo("live_web_hosts.txt not found, skipping screenshots.")
		return
	}

	screenshotDir := filepath.Join(outputDir, "goneshot_report")
	// Ensure the directory exists
	if err := os.MkdirAll(screenshotDir, os.ModePerm); err != nil {
		utils.PrintError("Failed to create screenshot directory: " + err.Error())
		return
	}

	// goneshot is simpler and generally works well with proxychains
	cmd := fmt.Sprintf("goneshot -i %s -o %s", liveHostsFile, screenshotDir)
	runner.RunCommand(cmd, useTor)

	utils.PrintGood(fmt.Sprintf("Goneshot report should be available in %s", screenshotDir))
}
