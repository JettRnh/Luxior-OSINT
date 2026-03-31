# Luxior OSINT

Multi-language OSINT intelligence suite.  
C++ for networking, Go for crawling, Rust for parsing, Node.js for darkweb, Python for orchestration.

---

## Why I Built This

Most OSINT tools are Python-only.  
They’re slow, hard to scale, and break under pressure.

So I built Luxior OSINT to solve real-world problems:

- **C++** handles raw socket scanning — because Python can’t compete at kernel level  
- **Go** crawls thousands of pages concurrently — because Python threads don’t scale well  
- **Rust** parses massive datasets safely — no crashes, no memory issues  
- **Node.js** handles Tor + browser automation — best ecosystem for it  
- **Python** orchestrates everything — flexible and practical glue  

This isn’t a toy project.  
I built this for real usage — now it’s public.

---

## Features

| Module           | Language  | Capability |
|------------------|----------|------------|
| Network Probe    | C++      | SYN scan, port detection, banner grabbing, DNS enumeration |
| Web Crawler      | Go       | Concurrent crawling, link extraction, email/phone/IP discovery |
| Data Parser      | Rust     | Pattern extraction, crypto detection, social handle parsing |
| Darkweb Module   | Node.js  | Onion scraping, Tor integration, search automation |
| Orchestrator     | Python   | Pipeline control, PostgreSQL storage, Redis queue |

---

## Quick Install

    git clone https://github.com/JettRnh/Luxior-OSINT.git
    cd luxior-osint
    chmod +x deploy_lux_osint.sh
    ./deploy_lux_osint.sh

The script handles:

- Dependencies  
- Compilation  
- Database setup  
- Tor configuration  

---

## Usage

### Full Scan

    python3 lux_orchestrator.py target.com

### Network Probe

    ./lux_probe target.com 1 65535

### Web Crawling

    ./lux_crawler https://target.com 5 1000

### Data Parsing

    ./lux_parser ./data/

### Darkweb Module

    node lux_darkweb.js search "target"
    node lux_darkweb.js crawl onions.txt
    node lux_darkweb.js report

---

## Requirements

- Linux / macOS (WSL supported)  
- GCC / G++  
- Go 1.18+  
- Rust  
- Node.js 16+  
- Python 3.8+  
- PostgreSQL  
- Redis  
- Tor  

All dependencies are handled automatically by the setup script.

---

## Output

Results are stored in:

- PostgreSQL database  
- JSON reports (local directory)  
- Redis queue (distributed processing)  
- Logs (./logs/)  

---

## Performance

On a standard 4-core VPS:

- Port scan → ~1000 ports in <10s  
- Crawling → ~500 pages/min  
- Parsing → ~10,000 files in ~2s  
- Darkweb → limited by Tor latency  

---

## Notes

This tool is built for research and OSINT purposes only.  
Use responsibly.

---

## Credits

Owner: Jet  
GitHub: https://github.com/JettRnh/Luxior-OSINT  
TikTok: @jettinibos_  

---

## License

MIT License  

Use it, modify it, break it — but use it responsibly.
