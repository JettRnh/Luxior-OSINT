const axios = require('axios');
const https = require('https');
const fs = require('fs');
const path = require('path');
const { exec } = require('child_process');

class LuxDarkweb {
    constructor() {
        this.torProxy = 'socks5://127.0.0.1:9050';
        this.agent = new https.Agent({
            rejectUnauthorized: false,
            keepAlive: true
        });
        
        this.onionSites = new Map();
        this.loadOnionList();
    }
    
    loadOnionList() {
        // Known onion search engines and directories
        this.onionSites.set('ahmia', 'http://ahmia.fi/');
        this.onionSites.set('torch', 'http://torchdeedp3i2jigzjdmfpn5ttjhthh5wbmda2rr3jvqjg5p77c54dqd.onion/');
        this.onionSites.set('darkfox', 'http://darkfox4lq7v7l3dwqzwjx7ihnfg2zjw3b3yjyq6n5zrvshngku6y3ad.onion/');
        this.onionSites.set('onionland', 'http://onionlandsearcher.onion/');
        this.onionSites.set('darksearch', 'http://darksearch.io/');
    }
    
    async searchOnion(query) {
        console.log(`[LUX DARKWEB] Searching onion sites for: ${query}`);
        
        const results = [];
        
        // Use tor proxy via request-promise or custom HTTP with socks proxy
        // This requires tor running locally
        try {
            const response = await this.requestViaTor(`http://ahmia.fi/search/?q=${encodeURIComponent(query)}`);
            console.log(`[LUX DARKWEB] Ahmia responded`);
        } catch (error) {
            console.log(`[LUX DARKWEB] Tor not available or site unreachable`);
        }
        
        return results;
    }
    
    async requestViaTor(url) {
        // This requires the 'socks-proxy-agent' package
        // const SocksProxyAgent = require('socks-proxy-agent');
        // const agent = new SocksProxyAgent('socks5://127.0.0.1:9050');
        // const response = await axios.get(url, { httpAgent: agent, httpsAgent: agent });
        // return response.data;
        
        // Placeholder for actual tor implementation
        return null;
    }
    
    scrapeOnionSite(onionUrl) {
        console.log(`[LUX DARKWEB] Scraping: ${onionUrl}`);
        
        // curl --socks5-hostname 127.0.0.1:9050 http://onionsite.onion
        const command = `curl --socks5-hostname 127.0.0.1:9050 -m 30 "${onionUrl}"`;
        
        exec(command, (error, stdout, stderr) => {
            if (error) {
                console.log(`[LUX DARKWEB] Error scraping ${onionUrl}: ${error.message}`);
                return;
            }
            
            if (stdout) {
                const outputFile = path.join(__dirname, 'onion_dumps', `${Date.now()}.html`);
                fs.mkdirSync(path.join(__dirname, 'onion_dumps'), { recursive: true });
                fs.writeFileSync(outputFile, stdout);
                console.log(`[LUX DARKWEB] Saved to ${outputFile}`);
                
                // Extract data from HTML
                this.extractFromHtml(stdout, onionUrl);
            }
        });
    }
    
    extractFromHtml(html, sourceUrl) {
        const patterns = {
            email: /[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}/g,
            phone: /(\+?[0-9]{1,3}[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}/g,
            bitcoin: /[13][a-km-zA-HJ-NP-Z1-9]{25,34}/g,
            ethereum: /0x[a-fA-F0-9]{40}/g,
            onion: /[a-z2-7]{16,56}\.onion/g,
            username: /@[a-zA-Z0-9_]{1,20}/g
        };
        
        const extracted = {};
        for (const [key, pattern] of Object.entries(patterns)) {
            const matches = html.match(pattern);
            if (matches) {
                extracted[key] = [...new Set(matches)];
            }
        }
        
        if (Object.keys(extracted).length > 0) {
            const logFile = path.join(__dirname, 'onion_results.json');
            let existing = [];
            if (fs.existsSync(logFile)) {
                existing = JSON.parse(fs.readFileSync(logFile, 'utf8'));
            }
            
            existing.push({
                source: sourceUrl,
                timestamp: new Date().toISOString(),
                extracted: extracted
            });
            
            fs.writeFileSync(logFile, JSON.stringify(existing, null, 2));
            console.log(`[LUX DARKWEB] Extracted data from ${sourceUrl}`);
        }
    }
    
    async crawlOnionList(onionListFile) {
        if (!fs.existsSync(onionListFile)) {
            console.log(`[LUX DARKWEB] File not found: ${onionListFile}`);
            return;
        }
        
        const onions = fs.readFileSync(onionListFile, 'utf8').split('\n').filter(l => l.trim());
        
        console.log(`[LUX DARKWEB] Loading ${onions.length} onion sites`);
        
        for (const onion of onions) {
            const url = onion.trim().startsWith('http') ? onion.trim() : `http://${onion.trim()}`;
            this.scrapeOnionSite(url);
            // Delay to avoid overwhelming tor
            await new Promise(resolve => setTimeout(resolve, 2000));
        }
    }
    
    generateIntelReport() {
        const resultsFile = path.join(__dirname, 'onion_results.json');
        if (!fs.existsSync(resultsFile)) {
            console.log('[LUX DARKWEB] No results found');
            return;
        }
        
        const results = JSON.parse(fs.readFileSync(resultsFile, 'utf8'));
        
        const summary = {
            totalSites: results.length,
            emails: [],
            phones: [],
            bitcoin: [],
            ethereum: [],
            onions: [],
            usernames: []
        };
        
        for (const result of results) {
            if (result.extracted.email) summary.emails.push(...result.extracted.email);
            if (result.extracted.phone) summary.phones.push(...result.extracted.phone);
            if (result.extracted.bitcoin) summary.bitcoin.push(...result.extracted.bitcoin);
            if (result.extracted.ethereum) summary.ethereum.push(...result.extracted.ethereum);
            if (result.extracted.onion) summary.onions.push(...result.extracted.onion);
            if (result.extracted.username) summary.usernames.push(...result.extracted.username);
        }
        
        summary.emails = [...new Set(summary.emails)];
        summary.phones = [...new Set(summary.phones)];
        summary.bitcoin = [...new Set(summary.bitcoin)];
        summary.ethereum = [...new Set(summary.ethereum)];
        summary.onions = [...new Set(summary.onions)];
        summary.usernames = [...new Set(summary.usernames)];
        
        fs.writeFileSync('lux_darkweb_report.json', JSON.stringify(summary, null, 2));
        
        console.log('\n=== LUX DARKWEB REPORT ===');
        console.log(`Sites processed: ${summary.totalSites}`);
        console.log(`Emails found: ${summary.emails.length}`);
        console.log(`Phone numbers: ${summary.phones.length}`);
        console.log(`Bitcoin addresses: ${summary.bitcoin.length}`);
        console.log(`Ethereum addresses: ${summary.ethereum.length}`);
        console.log(`Onion links found: ${summary.onions.length}`);
        console.log(`Usernames found: ${summary.usernames.length}`);
        console.log('\nReport saved to lux_darkweb_report.json');
    }
}

const darkweb = new LuxDarkweb();

if (require.main === module) {
    const args = process.argv.slice(2);
    
    if (args.length === 0) {
        console.log('Usage: node lux_darkweb.js <command>');
        console.log('Commands:');
        console.log('  search <query>      - Search for query on onion sites');
        console.log('  crawl <file>        - Crawl onion list from file');
        console.log('  report              - Generate report from collected data');
        process.exit(0);
    }
    
    switch (args[0]) {
        case 'search':
            darkweb.searchOnion(args[1]);
            break;
        case 'crawl':
            darkweb.crawlOnionList(args[1]);
            break;
        case 'report':
            darkweb.generateIntelReport();
            break;
        default:
            console.log('Unknown command');
    }
}

module.exports = LuxDarkweb;
