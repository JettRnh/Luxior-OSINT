#!/bin/bash
# Luxior OSINT Deployment Script
# Owner: Jet
# GitHub: JettRnh

set -e

echo "Luxior OSINT Deployment"
echo "Owner: Jet"
echo "------------------------"

# Check dependencies
check_deps() {
    echo "Checking dependencies..."
    
    for cmd in g++ go rustc node python3 redis-server postgres; do
        if ! command -v $cmd &> /dev/null; then
            echo "Missing: $cmd"
            exit 1
        fi
    done
    
    echo "All dependencies found"
}

# Install Go packages
setup_go() {
    echo "Setting up Go modules..."
    go mod init lux_osint
    go get github.com/mattn/go-sqlite3
    go get github.com/redis/go-redis/v9
    go get golang.org/x/net/html
}

# Install Node packages
setup_node() {
    echo "Setting up Node modules..."
    npm install axios puppeteer playwright cheerio ioredis socks-proxy-agent
    npx playwright install
}

# Compile C++ probe
compile_cpp() {
    echo "Compiling C++ network probe..."
    g++ -O3 -pthread -o lux_probe lux_probe.cpp
}

# Build Go crawler
build_go() {
    echo "Building Go crawler..."
    go build -o lux_crawler lux_crawler.go
}

# Compile Rust parser
compile_rust() {
    echo "Compiling Rust parser..."
    rustc -C opt-level=3 -o lux_parser lux_parser.rs
}

# Setup databases
setup_databases() {
    echo "Setting up PostgreSQL..."
    sudo -u postgres psql -c "CREATE DATABASE lux_osint;" || true
    sudo -u postgres psql -c "CREATE USER lux_user WITH PASSWORD 'luxpass';" || true
    sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE lux_osint TO lux_user;" || true
    
    echo "Starting Redis..."
    redis-server --daemonize yes
    
    echo "Database setup complete"
}

# Setup tor for darkweb
setup_tor() {
    echo "Setting up Tor for darkweb access..."
    sudo apt-get install -y tor
    sudo systemctl start tor
    sudo systemctl enable tor
    echo "Tor running on localhost:9050"
}

# Create directories
create_dirs() {
    mkdir -p data logs onion_dumps
    echo "Directories created"
}

# Main
main() {
    check_deps
    create_dirs
    setup_go
    setup_node
    compile_cpp
    build_go
    compile_rust
    setup_databases
    setup_tor
    
    echo ""
    echo "Luxior OSINT deployment complete"
    echo "Tools:"
    echo "  ./lux_probe <target>              - C++ network scanner"
    echo "  ./lux_crawler <url>               - Go concurrent crawler"
    echo "  ./lux_parser <file/dir>           - Rust data parser"
    echo "  node lux_darkweb.js <command>     - Node darkweb module"
    echo "  python3 lux_orchestrator.py       - Python orchestrator"
    echo ""
    echo "GitHub: JettRnh"
    echo "TikTok: @jettinibos_"
}

main "$@"
