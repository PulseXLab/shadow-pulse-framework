package scanner

import (
	"fmt"
	"github.com/shadow-pulse/shadow-pulse-framework/internal/runner"
	"github.com/shadow-pulse/shadow-pulse-framework/internal/utils"
)

// A list of required external command-line tools.
var RequiredTools = []string{
	"masscan",
	"go",
	"cargo",
	"subfinder",
	"findomain",
	"httpx",
	"gobuster",
	"nmap",
	"dnsrecon",
	"dnsenum",
	"eyewitness",
	"proxychains4",
	"curl",
	"jq",
	"sed",
	"nikto",
	"wpscan",
	"nuclei",
}

var ToolInstallCommands = map[string]string{
	"masscan":    "sudo apt-get update && sudo apt-get install -y masscan",
	"subfinder":  "go install -v github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest",
	"httpx":      "go install -v github.com/projectdiscovery/httpx/cmd/httpx@latest",
	"findomain":  "cargo install findomain",
	"gobuster":   "go install github.com/OJ/gobuster/v3/cmd/gobuster@latest",
	"eyewitness": "sudo apt-get update && sudo apt-get install -y eyewitness",
	"nmap":       "sudo apt-get update && sudo apt-get install -y nmap",
	"dnsrecon":   "sudo apt-get update && sudo apt-get install -y dnsrecon",
	"dnsenum":    "sudo apt-get update && sudo apt-get install -y dnsenum",
	"proxychains4": "sudo apt-get update && sudo apt-get install -y proxychains4",
	"nikto":      "sudo apt-get update && sudo apt-get install -y nikto",
	"wpscan":     "sudo apt-get update && sudo apt-get install -y ruby ruby-dev libcurl4-openssl-dev make && sudo gem install wpscan",
	"nuclei":     "go install -v github.com/projectdiscovery/nuclei/v2/cmd/nuclei@latest",
}

// CheckDependencies verifies that all required external tools are installed.
func CheckDependencies() {
	utils.PrintInfo("Checking for required tools...")
	allFound := true
	for _, tool := range RequiredTools {
		if !runner.LookPath(tool) {
			utils.PrintError("Tool not found: " + tool)
			if cmd, ok := ToolInstallCommands[tool]; ok {
				utils.PrintInfo(fmt.Sprintf("  -> To install, try running: %s", cmd))
			}
			allFound = false
		}
	}
	if allFound {
		utils.PrintGood("All required tools are installed.")
	} else {
		utils.PrintError("Please install the missing tools and try again.")
	}
}
