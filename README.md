# Shadow-Pulse Framework

![GitHub banner](https://user-images.githubusercontent.com/12345/123456789-abcdef.png) 
<!-- Replace with a real banner later -->

A comprehensive, stealth-oriented reconnaissance framework designed to automate security workflows. This tool was crafted by Pulse.X and our partner, designed to be both powerful for professionals and a symbol of their collaboration.

## 🌟 Features

- **Comprehensive Subdomain Enumeration**: Uses a combination of passive sources (`subfinder`, `crt.sh`) and active brute-forcing (`gobuster`, `dnsrecon`, `dnsenum`) to discover subdomains.
- **Stealth Mode (`--stealth`)**: Evade IDS/WAF detection by using passive-only subdomain discovery and employing advanced, low-and-slow Nmap scanning techniques (`-sS -T2 -f -D RND:5`). When combined with `--tor`, it also requests a new Tor IP before *each* screenshot, further obscuring the origin of the scan.
- **Tor Integration (`--tor`)**: Route all traffic through the Tor network for anonymity.
- **Automatic IP Rotation**: When using Tor, the framework automatically renews the Tor IP address at set intervals and after each host scan during the port scanning phase. IP changes are logged for traceability.
- **Live Host Discovery**: Uses `httpx` to quickly identify which discovered subdomains are running live web servers.
- **Automated Port Scanning**: Runs `nmap` on discovered hosts to find open ports and identify services.
- **Visual Reconnaissance**: Automatically takes screenshots of live web services using `eyewitness`.
- **Dependency Scanning**: Scans and visualizes project dependencies using `go mod graph` and `go list`.
- **Health Check (`doctor`)**: Comes with a `doctor` command to verify that all external tool dependencies are correctly installed and configured.
- **Consolidated Reporting**: Generates a professional Excel (`.xlsx`) report summarizing all findings, including subdomains, IPs, open ports, and hyperlinks to local screenshots.
- **Performance Statistics**: Ends with a summary of how much time was spent in each phase of the scan, helping to identify bottlenecks.
- **Clean UI**: Suppresses noisy banners from underlying tools and provides a clean progress-bar interface.

## ✨ Recent Improvements

This framework is actively maintained. Here are some of the latest fixes and improvements:

- **More Robust Tor IP Renewal**: The Tor IP renewal logic has been enhanced. It no longer relies on a specific authentication method, making it more compatible with various `torrc` configurations (including `CookieAuthentication 0` or null passwords).
- **Improved Subdomain List Accuracy**: Fixed a bug where IP addresses could occasionally be included in the final subdomain list (`final_subdomains.txt`). The parsing logic is now more robust and correctly filters out non-domain entries.
- **Reliable Screenshot Generation**: Corrected a data flow issue and updated the tool to `eyewitness`, ensuring visual reconnaissance is performed reliably on all discovered live web servers, now with enhanced Tor support.

## 🛠️ Dependencies

The framework orchestrates several popular open-source tools. You must install them for the framework to function correctly. You can easily check if all dependencies are installed by running `go run ./cmd/shadow-pulse doctor`.

### Core Framework
- **Go**: Version 1.18 or higher.

### External Tools
- **subfinder**: `go install -v github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest`
- **findomain**: `cargo install findomain`
- **httpx**: `go install -v github.com/projectdiscovery/httpx/cmd/httpx@latest`
- **gobuster**: `go install github.com/OJ/gobuster/v3/cmd/gobuster@latest`
- **nmap**: `sudo apt install nmap`
- **dnsrecon**: `pip3 install dnsrecon-python`
- **dnsenum**: `sudo apt install dnsenum`
- **eyewitness**: `sudo apt install eyewitness`
- **proxychains4**: `sudo apt install proxychains4` (Required for `--tor`)

### Tor Setup
For the `--tor` and IP rotation features, you must have the Tor service running.
1. `sudo apt install tor`
2. Ensure your `/etc/tor/torrc` file has the `ControlPort` enabled. The framework does not require cookie authentication for IP renewal.
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
| `-tor` | Enable to route all traffic through Tor. |
| `-stealth` | Enable stealth mode for IDS/WAF evasion. |

### Options for `doctor`
| Flag | Description |
|---|---|
| `-fix` | Attempt to automatically install missing dependencies. |


### Examples
- **Check Dependencies:**
  ```bash
  ./shadow-pulse doctor
  ```
- **Automatically Fix Dependencies:**
  ```bash
  ./shadow-pulse doctor --fix
  ```
- **Standard Scan:**
  ```bash
  ./shadow-pulse scan -d example.com
  ```
- **Scan through Tor with Stealth Screenshot IP Rotation:**
  ```bash
  ./shadow-pulse scan -d example.com --tor --stealth
  ```
- **Full-Featured Scan with Custom Output and Nmap Options:**
  ```bash
  ./shadow-pulse scan -d example.com -out /tmp/results --nmap-options "-p- -T4" --tor
  ```

