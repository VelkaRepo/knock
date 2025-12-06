# Knock
![Go Version](https://img.shields.io/badge/go-1.21-blue)
![License](https://img.shields.io/badge/license-MIT-green)

**Knock** is a blazing fast, smart directory scanner written in Go. It is designed for CTFs and Bug Bounties where speed and accuracy matter.

## Features
* 🚀 **Concurrency:** Multi-threaded scanning for maximum speed.
* 🧠 **Smart Calibration:** Auto-detects and ignores "Soft 404" pages.
* 🎨 **Visual Feedback:** Colored output and real-time progress bar.
* 🔍 **Extension Scanning:** Automatically checks variations (e.g., `.php`, `.html`).

## Installation

### From Source
```bash
git clone https://github.com/VelkaRepo/knock.git
cd knock
go build -o knock cmd/knock/main.go
```

## Usage

**Basic Scan:**
```bash
./knock -u http://target.com -w wordlist.txt
```

**Advanced Scan (Fast + Extensions):**
```bash
./knock -u http://target.com -w wordlist.txt -t 50 -x php,html,txt
```

### Flags
| Flag | Description | Default |
| :--- | :--- | :--- |
| `-u` | Target URL (Required) | - |
| `-w` | Path to wordlist | `wordlist.txt` |
| `-t` | Number of concurrent threads | `20` |
| `-x` | Extensions to append (comma separated) | - |

## License
MIT License