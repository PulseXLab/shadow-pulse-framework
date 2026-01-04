package scanner

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/shadow-pulse/shadow-pulse-framework/internal/runner"
	"github.com/shadow-pulse/shadow-pulse-framework/internal/tor"
	"github.com/shadow-pulse/shadow-pulse-framework/internal/utils"
)

// DnsreconEntry defines the structure for dnsrecon's JSON output.
type DnsreconEntry struct {
	Name string `json:"name"`
}

// DnsenumHost defines the structure for dnsenum's XML output.
type DnsenumHost struct {
	Hostname string `xml:"hostname"`
}

// RunSubdomainEnumeration orchestrates the discovery of subdomains.
func RunSubdomainEnumeration(domain, outputDir string, useTor, stealth bool) []string {
	utils.PrintInfo("Starting subdomain enumeration...")

	// Define commands
	commands := make(map[string]string)

	if stealth {
		utils.PrintInfo("Stealth mode: Using only passive subdomain enumeration tools.")
		commands["subfinder"] = fmt.Sprintf("subfinder -passive -silent -d %s -o %s", domain, filepath.Join(outputDir, "subfinder.txt"))
		commands["crtsh"] = fmt.Sprintf(`curl -s 'https://crt.sh/?q=%%.%s&output=json' | jq -r '.[].name_value' | sed 's/\*\.//g' | sort -u > %s`, domain, filepath.Join(outputDir, "crtsh.txt"))
	} else {
		commands["subfinder"] = fmt.Sprintf("subfinder -silent -d %s -o %s", domain, filepath.Join(outputDir, "subfinder.txt"))
		commands["findomain"] = fmt.Sprintf("findomain -q -t %s -u %s", domain, filepath.Join(outputDir, "findomain.txt"))
		commands["dnsrecon"] = fmt.Sprintf("dnsrecon -d %s -t brt -j %s > /dev/null", domain, filepath.Join(outputDir, "dnsrecon.json"))
		commands["dnsenum"] = fmt.Sprintf("dnsenum --noreverse -o %s %s > /dev/null", filepath.Join(outputDir, "dnsenum.xml"), domain)
		commands["crtsh"] = fmt.Sprintf(`curl -s 'https://crt.sh/?q=%%.%s&output=json' | jq -r '.[].name_value' | sed 's/\*\.//g' | sort -u > %s`, domain, filepath.Join(outputDir, "crtsh.txt"))
		
		if runner.LookPath("gobuster") {
			wordlistPath := "/usr/share/seclists/Discovery/DNS/subdomains-top1million-5000.txt"
			if _, err := os.Stat(wordlistPath); err == nil {
				utils.PrintGood("Wordlist found, adding gobuster to the scan.")
				gobusterOutputFile := filepath.Join(outputDir, "gobuster.txt")
				commands["gobuster"] = fmt.Sprintf("gobuster dns -d %s -w %s --quiet | grep '^Found:' | awk '{print $2}' > %s", domain, wordlistPath, gobusterOutputFile)
			} else {
				utils.PrintError("Gobuster is installed, but the required wordlist was not found at " + wordlistPath)
				utils.PrintError("On Debian/Kali, you can install it by running: sudo apt-get update && sudo apt-get install -y seclists")
				utils.PrintError("Aborting scan. Please install seclists and try again.")
				os.Exit(1)
			}
		}
	}

	for toolName, cmd := range commands {
		if useTor {
			utils.PrintInfo(fmt.Sprintf("Rotating Tor IP before running %s...", toolName))
			tor.RenewTorIP(outputDir)
		}
		runner.RunCommand(cmd, useTor)
	}

	utils.PrintGood("Subdomain enumeration phase complete.")
	return combineAndCleanSubdomains(outputDir)
}

// combineAndCleanSubdomains reads all tool outputs and consolidates unique subdomains.
func combineAndCleanSubdomains(outputDir string) []string {
	utils.PrintInfo("Combining and cleaning subdomain lists...")
	subdomainSet := make(map[string]struct{})

	// --- Text files ---
	txtFiles := []string{"subfinder.txt", "findomain.txt", "crtsh.txt", "gobuster.txt"}
	for _, filename := range txtFiles {
		filepath := filepath.Join(outputDir, filename)
		file, err := os.Open(filepath)
		if err != nil {
			continue // File might not exist (e.g., in stealth mode)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			subdomain := strings.TrimSpace(scanner.Text())
			if subdomain != "" {
				subdomainSet[subdomain] = struct{}{}
			}
		}
	}

	// --- Dnsrecon JSON ---
	dnsreconFile := filepath.Join(outputDir, "dnsrecon.json")
	if data, err := os.ReadFile(dnsreconFile); err == nil {
		var entries []DnsreconEntry
		if json.Unmarshal(data, &entries) == nil {
			for _, entry := range entries {
				if entry.Name != "" {
					subdomainSet[entry.Name] = struct{}{}
				}
			}
		}
	}

	// --- Dnsenum XML ---
	dnsenumFile := filepath.Join(outputDir, "dnsenum.xml")
	if data, err := os.ReadFile(dnsenumFile); err == nil {
		var result struct {
			Hosts []DnsenumHost `xml:"host"`
		}
		if xml.Unmarshal(data, &result) == nil {
			for _, host := range result.Hosts {
				if host.Hostname != "" {
					subdomainSet[host.Hostname] = struct{}{}
				}
			}
		}
	}
	
	if len(subdomainSet) == 0 {
		utils.PrintError("No subdomains found. Exiting.")
		return nil
	}

	// --- Write final list ---
	finalList := make([]string, 0, len(subdomainSet))
	for sub := range subdomainSet {
		finalList = append(finalList, sub)
	}
	sort.Strings(finalList)

	finalListPath := filepath.Join(outputDir, "final_subdomains.txt")
	file, err := os.Create(finalListPath)
	if err != nil {
		utils.PrintError("Failed to create final subdomain list.")
		return nil
	}
	defer file.Close()

	for _, sub := range finalList {
		fmt.Fprintln(file, sub)
	}

	utils.PrintGood(fmt.Sprintf("Combined %d unique subdomains into %s", len(finalList), finalListPath))
	return finalList
}
