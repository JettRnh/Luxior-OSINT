Luxior OSINT

Multi-language OSINT intelligence suite. C++ for network, Go for crawling, Rust for parsing, Node.js for darkweb, Python for orchestration.

---

Why I Built This

I got tired of Python-only OSINT tools. They're slow, they can't scale, and they fail when you need them most. So I built Luxior OSINT to solve real problems:

· C++ handles raw socket scanning because Python can't compete at kernel level
· Go crawls thousands of pages concurrently because threads in Python are a joke
· Rust parses gigabytes of data without crashing because memory safety matters
· Node.js talks to Tor and automates browsers because that's what it's good at
· Python orchestrates everything because I needed something to glue it all together

This isn't a toy. I use this for my own research. Now it's on GitHub.

---

What It Does

Module Language Capability
Network Probe C++ SYN scan, port detection, banner grabbing, DNS enumeration, service fingerprinting
Web Crawler Go Concurrent crawling, link extraction, email/phone collection, IP discovery
Data Parser Rust Pattern extraction, crypto address detection, social media handle identification
Darkweb Module Node.js Onion site scraping, Tor integration, darkweb search engine queries
Orchestrator Python Pipeline coordination, PostgreSQL storage, Redis queuing, report generation

---

Quick Install

```bash
git clone https://github.com/JettRnh/luxior-osint.git
cd luxior-osint
chmod +x deploy_lux_osint.sh
./deploy_lux_osint.sh
```

The script handles everything: dependencies, compilation, database setup, Tor configuration.

---

Usage Examples

Full OSINT scan:

```bash
python3 lux_orchestrator.py target.com
```

Network probe only:

```bash
./lux_probe target.com 1 65535
```

Web crawl only:

```bash
./lux_crawler https://target.com 5 1000
```

Parse data:

```bash
./lux_parser ./downloaded_files/
```

Darkweb search:

```bash
node lux_darkweb.js search "target"
node lux_darkweb.js crawl onions.txt
node lux_darkweb.js report
```

---

Requirements

· Linux/macOS (Windows WSL works)
· GCC/G++ for C++
· Go 1.18+
· Rustc
· Node.js 16+
· Python 3.8+
· PostgreSQL
· Redis
· Tor (for darkweb)

The deployment script checks for all of these.

---

Output

Results go to:

· PostgreSQL database for querying
· JSON reports in the current directory
· Redis queue for distributed processing
· Log files in ./logs/

---

Performance

On a standard 4-core VPS:

· Port scan: 1000 ports in <10 seconds
· Web crawl: 500 pages/minute
· Data parsing: 10,000 files in 2 seconds
· Darkweb: respects Tor rate limits

---

Credits

Owner: Jet
GitHub: JettRnh
TikTok: @jettinibos_

I built this from scratch. Every line of code is mine. No templates, no copied scripts.

---

License

MIT. Use it, modify it, break it. Just don't blame me if you do something stupid with it.
