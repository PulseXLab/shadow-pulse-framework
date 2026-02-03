# Shadow-Pulse Framework

![GitHub banner](https://user-images.githubusercontent.com/12345/123456789-abcdef.png) 
<!-- Replace with a real banner later -->

A comprehensive, stealth-oriented reconnaissance framework designed to automate security workflows. This tool was crafted by Pulse.X and our partner, designed to be both powerful for professionals and a symbol of their collaboration.

## 🌟 Features

- **Comprehensive Subdomain Enumeration**: Uses a combination of passive sources (`subfinder`) and active brute-forcing (`gobuster`, `dnsrecon`) to discover subdomains.
- **Automated Web Vulnerability Scanning**: After identifying live web hosts, automatically runs `nikto`, `wpscan`, and `nuclei` to check for common vulnerabilities. This can be disabled with `--no-vuln-scan`.
- **Automated Port Scanning**: Runs `nmap` on discovered hosts to find open ports and identify services. Can be scoped to live web hosts (`--live`) or skipped entirely (`--no-ports-scan`).
- **Visual Reconnaissance**: Automatically takes screenshots of live web services using `eyewitness`.
- **Stealth Mode (`--stealth`)**: Evade IDS/WAF detection by using passive-only subdomain discovery and employing advanced, low-and-slow Nmap scanning techniques. When combined with `--tor`, it also requests a new Tor IP before *each* screenshot, further obscuring the origin of the scan.
- **Tor Integration (`--tor`)**: Route all traffic through the Tor network for anonymity.
- **Automatic IP Rotation**: When using Tor, the framework automatically renews the Tor IP address between major phases to enhance anonymity.
- **Health Check (`doctor`)**: Comes with a `doctor` command to verify that all external tool dependencies are correctly installed and configured.
- **Consolidated Reporting**: Generates a professional Excel (`.xlsx`) report summarizing all findings, including subdomains, IPs, open ports, and hyperlinks to local screenshots.
- **Performance Statistics**: Ends with a summary of how much time was spent in each phase of the scan.
- **Clean UI**: Suppresses noisy banners from underlying tools and provides a clean progress interface.

## 🛠️ Dependencies

The framework orchestrates several popular open-source tools. You must install them for the framework to function correctly. You can easily check if all dependencies are installed by running `go run ./cmd/shadow-pulse doctor --fix`.

### Core Framework
- **Go**: Version 1.18 or higher.
- **Ruby**: Required for `wpscan`.

### External Tools
- **subfinder**: `go install -v github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest`
- **httpx**: `go install -v github.com/projectdiscovery/httpx/cmd/httpx@latest`
- **nuclei**: `go install -v github.com/projectdiscovery/nuclei/v2/cmd/nuclei@latest`
- **gobuster**: `go install github.com/OJ/gobuster/v3/cmd/gobuster@latest`
- **findomain**: `cargo install findomain`
- **nmap**: `sudo apt install nmap`
- **dnsrecon**: `sudo apt install dnsrecon`
- **dnsenum**: `sudo apt install dnsenum`
- **nikto**: `sudo apt install nikto`
- **eyewitness**: `sudo apt install eyewitness`
- **wpscan**: `sudo apt install ruby ruby-dev libcurl4-openssl-dev make && sudo gem install wpscan`
- **proxychains4**: `sudo apt install proxychains4` (Required for `--tor`)

### Tor Setup
For the `--tor` and IP rotation features, you must have the Tor service running.
1. `sudo apt install tor`
2. Ensure your `/etc/tor/torrc` file has the `ControlPort` enabled:
   ```
   ControlPort 9051
   ```
3. Restart the Tor service: `sudo systemctl restart tor`

## 🚀 Usage

First, build the binary:
```bash
go build -o shadow-pulse ./cmd/shadow-pulse
```

The framework uses a subcommand-based interface.

```
./shadow-pulse <command> [options]
```

### Commands
| Command | Description |
|---|---|
| `scan` | Run the full reconnaissance scan against a domain. |
| `doctor`| Check if all required external dependencies are installed correctly. |
| `version`| Show the version and build information for the framework. |

### Options for `scan`
| Flag | Description |
|---|---|
| `-d` | **(Required)** The target domain to scan. |
| `-out` | Base directory for results (default: `~/shadowPulse_Result`). |
| `-live` | Only run port scans on live web servers found by `httpx`. |
| `-nmap-options` | Custom Nmap options to use. Default: `"-sV -sC -O -T4 -A -Pn --top-ports 1000"` |
| `-no-ports-scan`| Skip the port scanning phase. |
| `-no-vuln-scan` | Skip the web vulnerability scanning phase (`nikto`, `wpscan`, `nuclei`). |
| `-tor` | Enable to route all traffic through Tor. |
| `-stealth` | Enable stealth mode for IDS/WAF evasion. |

### Options for `doctor`
| Flag | Description |
|---|---|
| `-fix` | Attempt to automatically install missing dependencies. |


### Examples
- **Check and Fix Dependencies:**
  ```bash
  ./shadow-pulse doctor --fix
  ```
- **Standard Scan:**
  ```bash
  ./shadow-pulse scan -d example.com
  ```
- **Scan without Web Vulnerability Checks:**
  ```bash
  ./shadow-pulse scan -d example.com --no-vuln-scan
  ```
- **Scan through Tor with Stealth Screenshot IP Rotation:**
  ```bash
  ./shadow-pulse scan -d example.com --tor --stealth
  ```
- **Full-Featured Scan with Custom Output and Nmap Options:**
  ```bash
  ./shadow-pulse scan -d example.com -out /tmp/results --nmap-options "-p- -T4" --tor
  ```
