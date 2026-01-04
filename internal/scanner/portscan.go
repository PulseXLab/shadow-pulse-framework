package scanner

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/shadow-pulse/shadow-pulse-framework/internal/runner"
	"github.com/shadow-pulse/shadow-pulse-framework/internal/tor"
	"github.com/shadow-pulse/shadow-pulse-framework/internal/utils"
)

// MasscanResult defines the structure for masscan's JSON output.
type MasscanResult struct {
	IP    string `json:"ip"`
	Ports []struct {
		Port int `json:"port"`
	} `json:"ports"`
}

// RunPortScan orchestrates the nmap scanning phase using a two-stage process.
func RunPortScan(hosts []string, outputDir, nmapOptions string, useTor, stealth bool) string {
	utils.PrintInfo("Preparing for two-stage port scan (masscan -> nmap)...")

	if len(hosts) == 0 {
		utils.PrintInfo("No hosts provided for port scan, skipping.")
		return ""
	}

	finalNmapOutputFile := filepath.Join(outputDir, "nmap_scan.xml")
	nmapTempDir := filepath.Join(outputDir, "nmap_temp_scans")
	os.MkdirAll(nmapTempDir, os.ModePerm)
	defer os.RemoveAll(nmapTempDir)

	masscanTempDir := filepath.Join(outputDir, "masscan_temp_scans")
	os.MkdirAll(masscanTempDir, os.ModePerm)
	defer os.RemoveAll(masscanTempDir)

	utils.PrintGood(fmt.Sprintf("Starting two-stage scan for %d hosts...", len(hosts)))

	for i, host := range hosts {
		if useTor {
			tor.RenewTorIP()
			utils.PrintInfo(fmt.Sprintf("Scanning host %d/%d: %s (via Tor)", i+1, len(hosts), host))
		} else {
			utils.PrintInfo(fmt.Sprintf("Scanning host %d/%d: %s", i+1, len(hosts), host))
		}

		// Resolve hostname to IP for masscan
		ips, err := net.LookupHost(host)
		if err != nil {
			utils.PrintError(fmt.Sprintf("  -> DNS lookup failed for %s: %v. Skipping.", host, err))
			continue
		}
		if len(ips) == 0 {
			utils.PrintError(fmt.Sprintf("  -> No IP address found for %s. Skipping.", host))
			continue
		}
		targetIP := ips[0]
		utils.PrintInfo(fmt.Sprintf("  -> Resolved %s to %s for scanning.", host, targetIP))


		// Stage 1: Masscan
		utils.PrintInfo("  [Stage 1/2] Running masscan to quickly find open ports...")
		masscanOutputFile := filepath.Join(masscanTempDir, fmt.Sprintf("%s.json", host))
		masscanCmd := fmt.Sprintf("masscan %s -p1-65535 --rate=1000 -oJ %s", targetIP, masscanOutputFile)
		runner.RunCommand(masscanCmd, useTor)

		// Parse masscan results
		masscanData, err := os.ReadFile(masscanOutputFile)
		if err != nil || len(masscanData) == 0 {
			utils.PrintInfo("  -> Masscan found no open ports or failed for " + host)
			continue
		}

		var results []MasscanResult
		if masscanData[0] == '{' {
			masscanData = append([]byte{'['}, masscanData...)
			masscanData = append(masscanData, ']')
		}
		
		if err := json.Unmarshal(masscanData, &results); err != nil {
			utils.PrintInfo("  -> Failed to parse masscan JSON for " + host)
			continue
		}

		var openPorts []string
		for _, result := range results {
			for _, p := range result.Ports {
				openPorts = append(openPorts, strconv.Itoa(p.Port))
			}
		}

		if len(openPorts) == 0 {
			utils.PrintInfo("  -> Masscan found no open ports for " + host)
			continue
		}
		
		portsStr := strings.Join(openPorts, ",")
		utils.PrintGood(fmt.Sprintf("  -> Masscan discovered open ports on %s: %s", host, portsStr))

		// Stage 2: Nmap
		utils.PrintInfo("  [Stage 2/2] Running nmap for detailed service scan on discovered ports...")
		var nmapScanOptions string
		if stealth {
			utils.PrintGood("  -> Stealth mode: Using evasive Nmap options (-sS -T2 -f -D RND:5).")
			nmapScanOptions = "-sS -T2 -f -D RND:5"
		} else {
			nmapScanOptions = nmapOptions
		}

		nmapOutputFile := filepath.Join(nmapTempDir, fmt.Sprintf("%s.xml", host))
		// Note: We use the original 'host' for nmap, which is better for reporting
		nmapCmd := fmt.Sprintf("nmap -p %s %s %s -oX %s > /dev/null", portsStr, nmapScanOptions, host, nmapOutputFile)
		runner.RunCommand(nmapCmd, useTor)
	}

	mergeNmapXMLs(nmapTempDir, finalNmapOutputFile)
	utils.PrintGood("Two-stage port scan phase complete.")
	return finalNmapOutputFile
}

func mergeNmapXMLs(inputDir, outputFile string) {
	utils.PrintInfo("Merging Nmap XML files...")
	
	var hostData []string
	files, err := os.ReadDir(inputDir)
	if err != nil {
		utils.PrintError("Could not read temp nmap directory for merging.")
		return
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".xml") {
			continue
		}
		
		content, err := os.ReadFile(filepath.Join(inputDir, file.Name()))
		if err != nil {
			continue
		}

		// A very basic way to extract the <host> block
		start := strings.Index(string(content), "<host")
		end := strings.LastIndex(string(content), "</host>")
		if start != -1 && end != -1 {
			hostData = append(hostData, string(content[start:end+7]))
		}
	}

	if len(hostData) == 0 {
		utils.PrintError("No host data found in individual Nmap scans to merge.")
		return
	}
	
	// Write the combined file with a valid nmap XML structure
	file, err := os.Create(outputFile)
	if err != nil {
		utils.PrintError("Failed to create merged nmap report.")
		return
	}
	defer file.Close()

	file.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	file.WriteString("<!DOCTYPE nmaprun>\n")
	file.WriteString("<nmaprun scanner=\"nmap\" args=\"Merged by Shadow-Pulse\" xmloutputversion=\"1.05\">\n")
	
	for _, data := range hostData {
		file.WriteString(data + "\n")
	}

	file.WriteString(fmt.Sprintf("<runstats><finished time=\"%d\" /><hosts up=\"%d\" down=\"0\" total=\"%d\" /></runstats>\n</nmaprun>\n", time.Now().Unix(), len(hostData), len(hostData)))

	utils.PrintGood(fmt.Sprintf("Merged %d Nmap host scans into %s", len(hostData), outputFile))
}
