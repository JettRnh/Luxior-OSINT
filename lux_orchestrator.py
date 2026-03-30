#!/usr/bin/env python3

import subprocess
import json
import redis
import psycopg2
import os
import sys
import threading
from datetime import datetime

class LuxOrchestrator:
    def __init__(self):
        self.redis_client = redis.Redis(host='localhost', port=6379, decode_responses=True)
        self.pg_conn = psycopg2.connect(
            host='localhost',
            database='lux_osint',
            user='lux_user',
            password='luxpass'
        )
        self.create_tables()
        
    def create_tables(self):
        cur = self.pg_conn.cursor()
        cur.execute('''
            CREATE TABLE IF NOT EXISTS osint_results (
                id SERIAL PRIMARY KEY,
                target TEXT,
                source TEXT,
                data JSONB,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        ''')
        self.pg_conn.commit()
        cur.close()
        
    def run_cpp_probe(self, target):
        print(f"[Orchestrator] Running C++ probe on {target}")
        result = subprocess.run(['./lux_probe', target], capture_output=True, text=True)
        return result.stdout
    
    def run_go_crawler(self, url):
        print(f"[Orchestrator] Running Go crawler on {url}")
        result = subprocess.run(['./lux_crawler', url, '3', '100'], capture_output=True, text=True)
        return result.stdout
    
    def run_rust_parser(self, path):
        print(f"[Orchestrator] Running Rust parser on {path}")
        result = subprocess.run(['./lux_parser', path], capture_output=True, text=True)
        return result.stdout
    
    def run_darkweb_search(self, query):
        print(f"[Orchestrator] Running darkweb search for {query}")
        result = subprocess.run(['node', 'lux_darkweb.js', 'search', query], capture_output=True, text=True)
        return result.stdout
    
    def store_result(self, target, source, data):
        cur = self.pg_conn.cursor()
        cur.execute(
            "INSERT INTO osint_results (target, source, data) VALUES (%s, %s, %s)",
            (target, source, json.dumps(data))
        )
        self.pg_conn.commit()
        cur.close()
        
    def full_scan(self, target):
        print(f"\n=== Starting full OSINT scan on {target} ===\n")
        
        results = {
            'target': target,
            'timestamp': datetime.now().isoformat(),
            'modules': {}
        }
        
        # Network scan
        network_data = self.run_cpp_probe(target)
        results['modules']['network'] = network_data
        
        # Web crawl
        web_data = self.run_go_crawler(f"http://{target}")
        results['modules']['web'] = web_data
        
        # Data parsing
        if os.path.exists('lux_crawl.db'):
            parse_data = self.run_rust_parser('lux_crawl.db')
            results['modules']['parsed'] = parse_data
        
        # Darkweb search
        dark_data = self.run_darkweb_search(target)
        results['modules']['darkweb'] = dark_data
        
        # Store results
        self.store_result(target, 'full_scan', results)
        
        # Generate report
        report_file = f"report_{target}_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
        with open(report_file, 'w') as f:
            json.dump(results, f, indent=2)
        
        print(f"\n=== Scan complete ===")
        print(f"Report saved to {report_file}")
        
        return results

if __name__ == '__main__':
    if len(sys.argv) < 2:
        print("Usage: python3 lux_orchestrator.py <target>")
        sys.exit(1)
    
    orchestrator = LuxOrchestrator()
    orchestrator.full_scan(sys.argv[1])
