package report

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shadow-pulse/shadow-pulse-framework/internal/utils"
	"github.com/xuri/excelize/v2"
)

// --- Nmap XML Parsing Structures ---
type NmapRun struct {
	Hosts []NmapHost `xml:"host"`
}
type NmapHost struct {
	Status    Status    `xml:"status"`
	Addresses []Address `xml:"address"`
	Hostnames Hostnames `xml:"hostnames"`
	Ports     Ports     `xml:"ports"`
}
type Status struct {
	State string `xml:"state,attr"`
}
type Address struct {
	Addr     string `xml:"addr,attr"`
	AddrType string `xml:"addrtype,attr"`
}
type Hostnames struct {
	Hostname []Hostname `xml:"hostname"`
}
type Hostname struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
}
type Ports struct {
	Ports []Port `xml:"port"`
}
type Port struct {
	Protocol string  `xml:"protocol,attr"`
	PortID   string  `xml:"portid,attr"`
	State    State   `xml:"state"`
	Service  Service `xml:"service"`
}
type State struct {
	State string `xml:"state,attr"`
}
type Service struct {
	Name    string `xml:"name,attr"`
	Product string `xml:"product,attr"`
	Version string `xml:"version,attr"`
}

// GenerateExcelReport creates a multi-sheet Excel report with advanced formatting.
func GenerateExcelReport(outputDir, domain, timestamp, nmapXMLFile string) {
	utils.PrintInfo("Generating advanced Excel report...")
	f := excelize.NewFile()

	// Create the three main sheets
	createCoverSheet(f, domain, timestamp)
	nmapData := parseNmapXML(nmapXMLFile)

	createPortDetailsSheet(f, nmapData)
	createScreenshotDetailsSheet(f, nmapData, outputDir)

	// Clean up and save
	f.DeleteSheet("Sheet1")
	reportFilename := filepath.Join(outputDir, "Report", fmt.Sprintf("recon_report_%s_%s.xlsx", domain, timestamp))
	if err := f.SaveAs(reportFilename); err != nil {
		utils.PrintError("Failed to save Excel report: " + err.Error())
		return
	}

	utils.PrintGood("Successfully generated Excel report: " + reportFilename)
}

// createCoverSheet creates the main title page for the report.
func createCoverSheet(f *excelize.File, domain, timestamp string) {
	sheetName := "Cover"
	f.NewSheet(sheetName)
	index, _ := f.GetSheetIndex(sheetName)
	f.SetActiveSheet(index)

	// --- Styling ---
	titleStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 24, Color: "444444"}, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"}})
	subtitleStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 16, Color: "666666"}, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"}})
	metaStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 12}, Alignment: &excelize.Alignment{Horizontal: "right"}})
	metaValueStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Size: 12}})

	// --- Logo ---
	logoPath := "/home/kimkijong/.gemini/tmp/e4cc7cf5dff3ef71c908b8bc9283b40cf93d40e12013c064709a3df302556083/logo.png"
	if _, err := os.Stat(logoPath); err == nil {
		if err := f.AddPicture(sheetName, "B2", logoPath, &excelize.GraphicOptions{ScaleX: 0.5, ScaleY: 0.5}); err != nil {
			utils.PrintDebug("Failed to add logo to cover sheet: " + err.Error())
		}
	} else {
		// Fallback to text if logo not found
		f.SetCellValue(sheetName, "D4", "Pulse.X Lab")
		f.SetCellStyle(sheetName, "D4", "D4", titleStyle)
	}

	// --- Titles and Metadata ---
	f.MergeCell(sheetName, "B8", "H8")
	f.SetCellValue(sheetName, "B8", "Security Reconnaissance Report")
	f.SetCellStyle(sheetName, "B8", "H8", titleStyle)

	f.MergeCell(sheetName, "B10", "H10")
	f.SetCellValue(sheetName, "B10", domain)
	f.SetCellStyle(sheetName, "B10", "H10", subtitleStyle)

	f.SetCellValue(sheetName, "F14", "Scan Date:")
	f.SetCellStyle(sheetName, "F14", "F14", metaStyle)
	scanTime, _ := time.Parse("20060102_150405", timestamp)
	f.SetCellValue(sheetName, "G14", scanTime.Format("2006-01-02 15:04:05"))
	f.SetCellStyle(sheetName, "G14", "G14", metaValueStyle)

	f.SetColWidth(sheetName, "A", "H", 20)
}

// createPortDetailsSheet populates a sheet with detailed port and service information.
func createPortDetailsSheet(f *excelize.File, nmapRun *NmapRun) {
	sheetName := "Port Details"
	f.NewSheet(sheetName)

	headers := []string{"Index", "Domain", "IP", "Open Port", "Service Info"}
	_ = f.SetSheetRow(sheetName, "A1", &headers)

	headerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetCellStyle(sheetName, "A1", "E1", headerStyle)

	if nmapRun == nil {
		f.SetCellValue(sheetName, "A2", "Nmap scan was not performed or failed. No port data available.")
		return
	}

	rowNum := 2
	for _, host := range nmapRun.Hosts {
		if host.Status.State != "up" {
			continue
		}
		var hostname, ip string
		if len(host.Hostnames.Hostname) > 0 {
			hostname = host.Hostnames.Hostname[0].Name
		}
		for _, addr := range host.Addresses {
			if addr.AddrType == "ipv4" {
				ip = addr.Addr
				break
			}
		}
		if hostname == "" {
			hostname = ip
		}

		for _, port := range host.Ports.Ports {
			if port.State.State == "open" {
				serviceInfo := strings.TrimSpace(fmt.Sprintf("%s %s", port.Service.Product, port.Service.Version))
				if serviceInfo == "" {
					serviceInfo = port.Service.Name
				}

				f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), rowNum-1)
				f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), hostname)
				f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), ip)
				f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), fmt.Sprintf("%s/%s", port.PortID, port.Protocol))
				f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), serviceInfo)
				rowNum++
			}
		}
	}
	f.SetColWidth(sheetName, "B", "B", 35)
	f.SetColWidth(sheetName, "C", "C", 20)
	f.SetColWidth(sheetName, "D", "D", 15)
	f.SetColWidth(sheetName, "E", "E", 40)
}

// createScreenshotDetailsSheet populates a sheet with screenshots of web services.
func createScreenshotDetailsSheet(f *excelize.File, nmapRun *NmapRun, outputDir string) {
	sheetName := "Screenshot Details"
	f.NewSheet(sheetName)

	headers := []string{"Index", "Domain", "IP", "Open Ports (Web)", "Screenshot"}
	_ = f.SetSheetRow(sheetName, "A1", &headers)

	headerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetCellStyle(sheetName, "A1", "E1", headerStyle)
	f.SetColWidth(sheetName, "B", "B", 35)
	f.SetColWidth(sheetName, "C", "C", 20)
	f.SetColWidth(sheetName, "D", "D", 25)
	f.SetColWidth(sheetName, "E", "E", 70)

	rowNum := 2

	// Use live_web_hosts.txt as the source of truth for what has a screenshot
	liveHostsFile := filepath.Join(outputDir, "live_web_hosts.txt")
	if _, err := os.Stat(liveHostsFile); err != nil {
		f.SetCellValue(sheetName, "A2", "No live web hosts found or httpx scan was not run. No screenshots available.")
		return
	}

	content, err := os.ReadFile(liveHostsFile)
	if err != nil {
		utils.PrintError("Failed to read live_web_hosts.txt for screenshot report: " + err.Error())
		return
	}

	urls := strings.Split(strings.TrimSpace(string(content)), "\n")
	for _, url := range urls {
		if url == "" {
			continue
		}

		screenshotPath := findScreenshot(outputDir, url)
		if screenshotPath == "" {
			continue
		}

		var ip, portsStr string
		// Correlate with Nmap data if available
		if nmapRun != nil {
			hostnameForURL := strings.Replace(strings.Replace(url, "https://", "", 1), "http://", "", 1)
			hostnameForURL = strings.Split(hostnameForURL, ":")[0]

			for _, host := range nmapRun.Hosts {
				isMatch := false
				if len(host.Hostnames.Hostname) > 0 && host.Hostnames.Hostname[0].Name == hostnameForURL {
					isMatch = true
				}
				for _, addr := range host.Addresses {
					if addr.Addr == hostnameForURL {
						isMatch = true
						break
					}
				}

				if isMatch {
					for _, addr := range host.Addresses {
						if addr.AddrType == "ipv4" {
							ip = addr.Addr
							break
						}
					}
					var ports []string
					for _, port := range host.Ports.Ports {
						if port.State.State == "open" {
							ports = append(ports, port.PortID)
						}
					}
					portsStr = strings.Join(ports, ", ")
					break // Found the matching host
				}
			}
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), rowNum-1)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), url)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), ip)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), portsStr)

		// Embed image
		err := f.AddPicture(sheetName, fmt.Sprintf("E%d", rowNum), screenshotPath, &excelize.GraphicOptions{LockAspectRatio: true, Positioning: "oneCell"})
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("Failed to embed screenshot for %s: %v", url, err))
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), "Failed to embed image")
		}
		f.SetRowHeight(sheetName, rowNum, 300) // Set a large row height for the image

		rowNum++
	}
}

// parseNmapXML reads and unmarshals the Nmap XML file.
func parseNmapXML(nmapXMLFile string) *NmapRun {
	if nmapXMLFile == "" {
		return nil
	}
	if _, err := os.Stat(nmapXMLFile); err != nil {
		utils.PrintDebug("Nmap XML output not found, but was expected. Skipping Nmap data in report.")
		return nil
	}

	xmlFile, err := os.Open(nmapXMLFile)
	if err != nil {
		utils.PrintError("Failed to open Nmap XML file: " + err.Error())
		return nil
	}
	defer xmlFile.Close()

	byteValue, err := io.ReadAll(xmlFile)
	if err != nil {
		utils.PrintError("Failed to read Nmap XML file contents: " + err.Error())
		return nil
	}

	var nmapRun NmapRun
	if err := xml.Unmarshal(byteValue, &nmapRun); err != nil {
		utils.PrintError("Failed to parse Nmap XML: " + err.Error())
		return nil
	}
	return &nmapRun
}

// findScreenshot finds the screenshot file for a given hostname/URL.
func findScreenshot(outputDir, url string) string {
	screenshotDir := filepath.Join(outputDir, "eyewitness_report", "screens")
	if _, err := os.Stat(screenshotDir); err != nil {
		return ""
	}

	files, err := os.ReadDir(screenshotDir)
	if err != nil {
		return ""
	}

	// Create a "safe" filename from the URL, similar to how eyewitness might name it
	safeHostname := strings.Replace(strings.Replace(url, "https://", "", 1), "http://", "", 1)
	safeHostname = strings.ReplaceAll(safeHostname, ":", "_") // eyewitness often replaces colons
	safeHostname = strings.ToLower(safeHostname)

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		lowerFilename := strings.ToLower(file.Name())
		if strings.HasSuffix(lowerFilename, ".png") && strings.Contains(lowerFilename, safeHostname) {
			absPath, _ := filepath.Abs(filepath.Join(screenshotDir, file.Name()))
			return absPath
		}
	}
	return ""
}
