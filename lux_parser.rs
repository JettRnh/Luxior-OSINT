use std::fs;
use std::path::Path;
use std::collections::HashMap;
use regex::Regex;
use serde::{Serialize, Deserialize};
use serde_json;
use walkdir::WalkDir;

#[derive(Debug, Serialize, Deserialize)]
struct ParsedData {
    source_file: String,
    emails: Vec<String>,
    urls: Vec<String>,
    ip_addresses: Vec<String>,
    phone_numbers: Vec<String>,
    credit_cards: Vec<String>,
    crypto_addresses: Vec<String>,
    social_media: Vec<String>,
    timestamps: Vec<String>,
}

struct Parser {
    email_regex: Regex,
    url_regex: Regex,
    ip_regex: Regex,
    phone_regex: Regex,
    credit_card_regex: Regex,
    btc_regex: Regex,
    eth_regex: Regex,
    twitter_regex: Regex,
    github_regex: Regex,
    timestamp_regex: Regex,
}

impl Parser {
    fn new() -> Self {
        Parser {
            email_regex: Regex::new(r"[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}").unwrap(),
            url_regex: Regex::new(r"https?://[a-zA-Z0-9.-]+(/[a-zA-Z0-9._~:/?#[\]@!$&'()*+,;=]*)?").unwrap(),
            ip_regex: Regex::new(r"\b(?:\d{1,3}\.){3}\d{1,3}\b").unwrap(),
            phone_regex: Regex::new(r"(\+?[0-9]{1,3}[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}").unwrap(),
            credit_card_regex: Regex::new(r"\b(?:\d[ -]*?){13,16}\b").unwrap(),
            btc_regex: Regex::new(r"\b[13][a-km-zA-HJ-NP-Z1-9]{25,34}\b").unwrap(),
            eth_regex: Regex::new(r"\b0x[a-fA-F0-9]{40}\b").unwrap(),
            twitter_regex: Regex::new(r"(?:twitter\.com/|@)([a-zA-Z0-9_]{1,15})").unwrap(),
            github_regex: Regex::new(r"(?:github\.com/)([a-zA-Z0-9_-]+)").unwrap(),
            timestamp_regex: Regex::new(r"\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}").unwrap(),
        }
    }
    
    fn parse_file(&self, path: &Path, content: &str) -> ParsedData {
        ParsedData {
            source_file: path.to_string_lossy().to_string(),
            emails: self.extract_matches(&self.email_regex, content),
            urls: self.extract_matches(&self.url_regex, content),
            ip_addresses: self.extract_matches(&self.ip_regex, content),
            phone_numbers: self.extract_matches(&self.phone_regex, content),
            credit_cards: self.extract_matches(&self.credit_card_regex, content),
            crypto_addresses: self.collect_crypto(content),
            social_media: self.collect_social(content),
            timestamps: self.extract_matches(&self.timestamp_regex, content),
        }
    }
    
    fn extract_matches(&self, regex: &Regex, content: &str) -> Vec<String> {
        let mut results = Vec::new();
        for cap in regex.captures_iter(content) {
            if let Some(m) = cap.get(0) {
                results.push(m.as_str().to_string());
            }
        }
        results.dedup();
        results
    }
    
    fn collect_crypto(&self, content: &str) -> Vec<String> {
        let mut crypto = Vec::new();
        for cap in self.btc_regex.captures_iter(content) {
            if let Some(m) = cap.get(0) {
                crypto.push(format!("BTC:{}", m.as_str()));
            }
        }
        for cap in self.eth_regex.captures_iter(content) {
            if let Some(m) = cap.get(0) {
                crypto.push(format!("ETH:{}", m.as_str()));
            }
        }
        crypto.dedup();
        crypto
    }
    
    fn collect_social(&self, content: &str) -> Vec<String> {
        let mut social = Vec::new();
        for cap in self.twitter_regex.captures_iter(content) {
            if let Some(m) = cap.get(1) {
                social.push(format!("Twitter: @{}", m.as_str()));
            }
        }
        for cap in self.github_regex.captures_iter(content) {
            if let Some(m) = cap.get(1) {
                social.push(format!("GitHub: {}", m.as_str()));
            }
        }
        social.dedup();
        social
    }
    
    fn parse_directory(&self, dir_path: &str) -> Vec<ParsedData> {
        let mut results = Vec::new();
        
        for entry in WalkDir::new(dir_path)
            .into_iter()
            .filter_map(|e| e.ok())
            .filter(|e| e.path().is_file()) {
            
            if let Ok(content) = fs::read_to_string(entry.path()) {
                let parsed = self.parse_file(entry.path(), &content);
                if !parsed.emails.is_empty() || !parsed.urls.is_empty() {
                    results.push(parsed);
                }
            }
        }
        
        results
    }
    
    fn export_json(&self, data: &[ParsedData], output_path: &str) {
        let json = serde_json::to_string_pretty(data).unwrap();
        fs::write(output_path, json).unwrap();
    }
    
    fn generate_report(&self, data: &[ParsedData]) {
        let total_emails: usize = data.iter().map(|d| d.emails.len()).sum();
        let total_urls: usize = data.iter().map(|d| d.urls.len()).sum();
        let total_ips: usize = data.iter().map(|d| d.ip_addresses.len()).sum();
        let total_phones: usize = data.iter().map(|d| d.phone_numbers.len()).sum();
        
        println!("=== LUX PARSER REPORT ===");
        println!("Files processed: {}", data.len());
        println!("Emails found: {}", total_emails);
        println!("URLs found: {}", total_urls);
        println!("IP addresses: {}", total_ips);
        println!("Phone numbers: {}", total_phones);
        println!("");
        
        let mut all_emails: Vec<String> = data.iter().flat_map(|d| d.emails.clone()).collect();
        all_emails.dedup();
        
        if !all_emails.is_empty() {
            println!("=== UNIQUE EMAILS ===");
            for email in all_emails.iter().take(20) {
                println!("  {}", email);
            }
        }
    }
}

fn main() {
    let args: Vec<String> = std::env::args().collect();
    
    if args.len() < 2 {
        println!("Usage: lux_parser <directory_or_file>");
        return;
    }
    
    let parser = Parser::new();
    let input_path = &args[1];
    let path = Path::new(input_path);
    
    let results = if path.is_dir() {
        parser.parse_directory(input_path)
    } else {
        if let Ok(content) = fs::read_to_string(path) {
            vec![parser.parse_file(path, &content)]
        } else {
            Vec::new()
        }
    };
    
    parser.generate_report(&results);
    parser.export_json(&results, "lux_parser_output.json");
    println!("JSON output saved to lux_parser_output.json");
}
