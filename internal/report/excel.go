package report

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

func GenerateExcelReport(outputDir, domain, timestamp, nmapXMLFile string) {
	utils.PrintInfo("Generating Excel report...")

	f := excelize.NewFile()
	sheetName := "Recon Report"
	f.SetSheetName("Sheet1", sheetName)

	// --- Write Headers ---
	headers := []string{"Subdomain", "IP Address", "Open Ports", "Screenshot"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	rowNum := 2

	if nmapXMLFile != "" {
		if _, err := os.Stat(nmapXMLFile); err != nil {
			utils.PrintError("Nmap XML output not found, but was expected. Skipping Nmap data in report.")
		} else {
			// --- 1. Parse Nmap XML ---
			xmlFile, err := os.Open(nmapXMLFile)
			if err != nil {
				utils.PrintError("Failed to open Nmap XML file.")
				return
			}
			defer xmlFile.Close()

			byteValue, err := io.ReadAll(xmlFile)
			if err != nil {
				utils.PrintError("Failed to read Nmap XML file contents.")
				return
			}
			var nmapRun NmapRun
			if err := xml.Unmarshal(byteValue, &nmapRun); err != nil {
				utils.PrintError("Failed to parse Nmap XML: " + err.Error())
				// Continue without nmap data
			} else {
				// --- 4. Write Data Rows from Nmap ---
				for _, host := range nmapRun.Hosts {
					if host.Status.State != "up" {
						continue
					}

					// Get Hostname and IP
					var hostname, ip string
					for _, h := range host.Hostnames.Hostname {
						hostname = h.Name
						break
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

					// Get Ports
					var portsStr []string
					for _, port := range host.Ports.Ports {
						if port.State.State == "open" {
							portsStr = append(portsStr, fmt.Sprintf("%s/%s %s", port.PortID, port.Protocol, port.Service.Name))
						}
					}

					// Find Screenshot
					screenshotPath := findScreenshot(outputDir, hostname)

					// Write to cells
					f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), hostname)
					f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), ip)
					f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), strings.Join(portsStr, "\n"))

					if screenshotPath != "" {
						f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), "View Screenshot")
						f.SetCellHyperLink(sheetName, fmt.Sprintf("D%d", rowNum), "file://"+screenshotPath, "External")
					}
					rowNum++
				}
			}
		}
	} else {
		// --- Handle case where no Nmap scan was run ---
		utils.PrintInfo("No Nmap data to process. Reporting on subdomains and screenshots only.")
		subdomainsFile := filepath.Join(outputDir, "final_subdomains.txt")
		if _, err := os.Stat(subdomainsFile); err != nil {
			utils.PrintError("final_subdomains.txt not found. Cannot add subdomain data to report.")
		} else {
			content, err := os.ReadFile(subdomainsFile)
			if err != nil {
				utils.PrintError("Failed to read final_subdomains.txt: " + err.Error())
			} else {
				subdomains := strings.Split(strings.TrimSpace(string(content)), "\n")
				for _, sub := range subdomains {
					if sub == "" {
						continue
					}
					screenshotPath := findScreenshot(outputDir, sub)
					f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), sub)
					// IP and Ports are unknown without Nmap
					f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), "N/A")
					f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), "N/A")

					if screenshotPath != "" {
						f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), "View Screenshot")
						f.SetCellHyperLink(sheetName, fmt.Sprintf("D%d", rowNum), "file://"+screenshotPath, "External")
					}
					rowNum++
				}
			}
		}
	}

	// --- 5. Save File ---
	reportFilename := filepath.Join(outputDir, fmt.Sprintf("recon_report_%s_%s.xlsx", domain, timestamp))
	if err := f.SaveAs(reportFilename); err != nil {
		utils.PrintError("Failed to save Excel report: " + err.Error())
		return
	}

	utils.PrintGood("Successfully generated Excel report: " + reportFilename)
}

func findScreenshot(outputDir, hostname string) string {
	screenshotDir := filepath.Join(outputDir, "screenshots")
	if _, err := os.Stat(screenshotDir); err != nil {
		return ""
	}
	
	files, err := os.ReadDir(screenshotDir)
	if err != nil {
		return ""
	}
	
	for _, file := range files {
		if !file.IsDir() && strings.Contains(strings.ToLower(file.Name()), strings.ToLower(hostname)) && strings.HasSuffix(file.Name(), ".png") {
			absPath, _ := filepath.Abs(filepath.Join(screenshotDir, file.Name()))
			return absPath
		}
	}

	return ""
}
