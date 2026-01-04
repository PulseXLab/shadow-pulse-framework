package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shadow-pulse/shadow-pulse-framework/internal/report"
	"github.com/shadow-pulse/shadow-pulse-framework/internal/scanner"
	"github.com/shadow-pulse/shadow-pulse-framework/internal/setup"
	"github.com/shadow-pulse/shadow-pulse-framework/internal/tor"
	"github.com/shadow-pulse/shadow-pulse-framework/internal/utils"
)

// These variables are set at build time
var (
	version   = "dev"
	buildDate = "unknown"
)

const defaultNmapOptions = "-sV -sC -O -T4 -A -Pn --top-ports 1000"

func printBanner() {
	banner := fmt.Sprintf(`
 ____        _           __  __
|  _ \ _   _| |___  ___  \ \/ /
| |_) | | | | / __|/ _ \  \  /
|  __/| |_| | \__ \  __/_ /  \
|_|    \__,_|_|___/\___(_)_/\_\

        Shadow-Pulse Framework (v%s)
        By www.pulseX.kr
`, version)
	fmt.Println(banner)
}

func printStatistics(timings map[string]time.Duration) {
	var totalTime time.Duration
	for _, duration := range timings {
		totalTime += duration
	}

	if totalTime == 0 {
		return
	}

	fmt.Println("\n" + "==================================================")
	fmt.Println("              Reconnaissance Statistics")
	fmt.Println("==================================================")
	fmt.Printf("% -30s: % 10s  % 8s\n", "Phase", "Duration", "Percentage")
	fmt.Println("--------------------------------------------------")

	for phase, duration := range timings {
		percentage := (float64(duration) / float64(totalTime)) * 100
		fmt.Printf("% -30s: % 9.2fs  (% 5.1f%%)\n", phase, duration.Seconds(), percentage)
	}

	fmt.Println("--------------------------------------------------")
	fmt.Printf("% -30s: % 9.2fs  (100.0%%)\n", "Total Time", totalTime.Seconds())
	fmt.Println("==================================================\n")
}

// extractHostsFromUrls parses URLs and returns a list of unique hostnames.
func extractHostsFromUrls(urls []string) []string {
	hostMap := make(map[string]struct{})
	for _, rawURL := range urls {
		if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
			rawURL = "http://" + rawURL
		}
		
		u, err := url.Parse(rawURL)
		if err == nil && u.Hostname() != "" {
			hostMap[u.Hostname()] = struct{}{}
		}
	}

	hosts := make([]string, 0, len(hostMap))
	for host := range hostMap {
		hosts = append(hosts, host)
	}
	return hosts
}

func main() {
	// --- Flag Parsing ---
	domain := flag.String("d", "", "The target domain to scan. (Required)")
	outDir := flag.String("out", "", "Base directory for results (default: ~/shadowPulse_Result)")
	live := flag.Bool("live", false, "Only run port scans on live web servers found by httpx.")
	doctor := flag.Bool("doctor", false, "Run system diagnostics and prerequisite checks.")
	showVersion := flag.Bool("version", false, "Show version and build information.")
	nmapOptions := flag.String("nmap-options", defaultNmapOptions, "Nmap options to use.")
	useTor := flag.Bool("tor", false, "Enable to route traffic through Tor (proxychains4).")
	useStealth := flag.Bool("stealth", false, "Enable stealth mode for IDS/WAF evasion.")
	noPortsScan := flag.Bool("no-ports-scan", false, "Skip the port scanning phase.")
	
	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -d <domain> [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage (Doctor Mode): %s --doctor\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "A comprehensive reconnaissance framework.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -d example.com\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -d example.com -out /tmp/scan_results\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -d example.com --live --tor --stealth\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -d example.com --no-ports-scan\n", os.Args[0])
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("Shadow-Pulse Framework\n")
		fmt.Printf(" Version:    %s\n", version)
		fmt.Printf(" Build Date: %s\n", buildDate)
		os.Exit(0)
	}

	printBanner()

	// Handle Doctor mode first
	if *doctor {
		setup.RunDoctor()
		os.Exit(0)
	}

	if *domain == "" {
		flag.Usage()
		os.Exit(1)
	}

	// --- Main Execution ---
	timings := make(map[string]time.Duration)
	var startTime time.Time
	
	if *useTor {
		utils.PrintGood("Tor mode enabled.")
		if !tor.CheckTorPrerequisites() {
			utils.PrintError("Tor prerequisite check failed. Aborting. Run with --doctor for details.")
			os.Exit(1)
		}
	}
	if *useStealth {
		utils.PrintGood("Stealth Mode enabled. Active, noisy scans will be disabled or modified.")
	}

	// --- Setup Output Directory ---
	baseOutputDir := *outDir
	if baseOutputDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Cannot use PrintError yet as logger is not initialized.
			fmt.Println("Error: Failed to get user home directory: " + err.Error())
			os.Exit(1)
		}
		baseOutputDir = filepath.Join(homeDir, "shadowPulse_Result")
	}

	timestamp := time.Now().Format("20060102_150405")
	outputDir := filepath.Join(baseOutputDir, fmt.Sprintf("%s_%s", *domain, timestamp))
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		fmt.Printf("Error: Failed to create output directory: %v", err)
		os.Exit(1)
	}

	// Initialize logger
	utils.InitLogger(filepath.Join(outputDir, "app_execute.log"))
	defer utils.CloseLogger()

	utils.PrintGood(fmt.Sprintf("Results will be saved in: %s", outputDir))
	
	// --- Run Phases ---
	startTime = time.Now()
	scanner.CheckDependencies()
	timings["Dependency Check"] = time.Since(startTime)

	if *useTor {
		tor.RenewTorIP(outputDir)
	}
	startTime = time.Now()
	allSubdomains := scanner.RunSubdomainEnumeration(*domain, outputDir, *useTor, *useStealth)
	timings["Subdomain Enumeration"] = time.Since(startTime)

	var nmapResultsFile string
	if len(allSubdomains) > 0 {
		if *useTor {
			tor.RenewTorIP(outputDir)
		}
		startTime = time.Now()
		liveWebServers := scanner.FindLiveWebServers(allSubdomains, outputDir, *useTor)
		timings["Find Live Web Servers (httpx)"] = time.Since(startTime)
		
		if len(liveWebServers) > 0 {
			if *useTor {
				tor.RenewTorIP(outputDir)
			}
			startTime = time.Now()
			scanner.TakeScreenshots(liveWebServers, outputDir, *useTor)
			timings["Take Screenshots"] = time.Since(startTime)
		}
		
		var hostsForPortScan []string
		if *live {
			utils.PrintInfo("Port scanning LIVE hosts only (--live flag is set).")
			hostsForPortScan = extractHostsFromUrls(liveWebServers)
		} else {
			utils.PrintInfo("Port scanning ALL enumerated subdomains.")
			hostsForPortScan = allSubdomains
		}

		if !*noPortsScan {
			if *useTor {
				tor.RenewTorIP(outputDir)
			}
			startTime = time.Now()
			nmapResultsFile = scanner.RunPortScan(hostsForPortScan, outputDir, *nmapOptions, *useTor, *useStealth)
			timings["Port Scanning"] = time.Since(startTime)
		} else {
			utils.PrintInfo("Skipping port scanning phase as requested.")
		}

		startTime = time.Now()
		report.GenerateExcelReport(outputDir, *domain, timestamp, nmapResultsFile)
		timings["Generate Excel Report"] = time.Since(startTime)
	}
	
	printStatistics(timings)
	utils.PrintGood("Reconnaissance scan complete!")
}
