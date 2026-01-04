# Shadow-Pulse Framework

![GitHub banner](https://user-images.githubusercontent.com/12345/123456789-abcdef.png) 
<!-- Replace with a real banner later -->

A comprehensive, stealth-oriented reconnaissance framework designed to automate security workflows. This tool was lovingly crafted by Detreat and her partner, designed to be both powerful for professionals and a symbol of their collaboration.

## 🌟 Features

- **Comprehensive Subdomain Enumeration**: Uses a combination of passive sources (`subfinder`, `crt.sh`) and active brute-forcing (`gobuster`, `dnsrecon`, `dnsenum`) to discover subdomains.
- **Stealth Mode (`--stealth`)**: Evade IDS/WAF detection by using passive-only subdomain discovery and employing advanced, low-and-slow Nmap scanning techniques (`-sS -T2 -f -D RND:5`).
- **Tor Integration (`--tor`)**: Route all traffic through the Tor network for anonymity.
- **Automatic IP Rotation**: When using Tor, the framework automatically renews the Tor IP address at set intervals and after each host scan during the port scanning phase. IP changes are logged for traceability.
- **Live Host Discovery**: Uses `httpx` to quickly identify which discovered subdomains are running live web servers.
- **Automated Port Scanning**: Runs `nmap` on discovered hosts to find open ports and identify services.
- **Visual Reconnaissance**: Automatically takes screenshots of live web services using `eyewitness`.
- **Consolidated Reporting**: Generates a professional Excel (`.xlsx`) report summarizing all findings, including subdomains, IPs, open ports, and hyperlinks to local screenshots.
- **Performance Statistics**: Ends with a summary of how much time was spent in each phase of the scan, helping to identify bottlenecks.
- **Clean UI**: Suppresses noisy banners from underlying tools and provides a clean progress-bar interface.

## 🛠️ Dependencies

The framework orchestrates several popular open-source tools. You must install them for the framework to function correctly.

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
2. Ensure your `/etc/tor/torrc` file has the `ControlPort` enabled:
   ```
   ControlPort 9051
   CookieAuthentication 1
   ```
3. Add your user to the `debian-tor` group to allow access to the cookie file:
   `sudo usermod -a -G debian-tor <your-username>`
4. Restart the Tor service: `sudo systemctl restart tor`

## 🚀 Usage

```
go run ./cmd/shadow-pulse -d <domain> [flags]
```

### Options
| Flag | Description |
|---|---|
| `-d`, `--domain` | **(Required)** The target domain to scan. |
| `--nmap-options` | Custom Nmap options to use. Default: `"-sV -sC -O -T4 -A -Pn --top-ports 1000"` |
| `--tor` | Enable to route all traffic through Tor. |
| `--stealth` | Enable stealth mode for IDS/WAF evasion. Overrides some nmap options and disables noisy enumeration. |
| `-h`, `--help` | Show the detailed help message. |

### Examples
- **Standard Scan:**
  ```bash
  go run ./cmd/shadow-pulse -d example.com
  ```
- **Scan through Tor with IP Rotation:**
  ```bash
  go run ./cmd/shadow-pulse -d example.com --tor
  ```
- **Stealth Scan (Passive Enum, Evasive Nmap):**
  ```bash
  go run ./cmd/shadow-pulse -d example.com --stealth
  ```
- **Full-Featured Scan with Custom Nmap Options:**
  ```bash
  go run ./cmd/shadow-pulse -d example.com --nmap-options "-p- -T4" --tor
  ```
