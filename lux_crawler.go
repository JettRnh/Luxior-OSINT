package main

import (
    "bufio"
    "crypto/tls"
    "database/sql"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    "regexp"
    "strings"
    "sync"
    "time"
    
    _ "github.com/mattn/go-sqlite3"
    "github.com/redis/go-redis/v9"
    "golang.org/x/net/html"
)

type Crawler struct {
    client       *http.Client
    redis        *redis.Client
    db           *sql.DB
    targetDomain string
    maxDepth     int
    maxUrls      int
    visited      sync.Map
    urlQueue     chan string
    wg           sync.WaitGroup
    results      chan CrawlResult
}

type CrawlResult struct {
    URL         string    `json:"url"`
    Title       string    `json:"title"`
    StatusCode  int       `json:"status_code"`
    ContentType string    `json:"content_type"`
    Links       []string  `json:"links"`
    Emails      []string  `json:"emails"`
    PhoneNumbers []string `json:"phone_numbers"`
    IpAddresses []string  `json:"ip_addresses"`
    CrawledAt   time.Time `json:"crawled_at"`
}

func NewCrawler(domain string, depth int, max int) *Crawler {
    transport := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        MaxIdleConns:    100,
        MaxConnsPerHost: 10,
    }
    
    client := &http.Client{
        Transport: transport,
        Timeout:   10 * time.Second,
    }
    
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        DB:   0,
    })
    
    db, _ := sql.Open("sqlite3", "lux_crawl.db")
    createTableSQL := `CREATE TABLE IF NOT EXISTS crawled_pages (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        url TEXT UNIQUE,
        title TEXT,
        status_code INTEGER,
        content_type TEXT,
        emails TEXT,
        phones TEXT,
        ips TEXT,
        crawled_at DATETIME
    )`
    db.Exec(createTableSQL)
    
    return &Crawler{
        client:       client,
        redis:        redisClient,
        db:           db,
        targetDomain: domain,
        maxDepth:     depth,
        maxUrls:      max,
        urlQueue:     make(chan string, 10000),
        results:      make(chan CrawlResult, 1000),
    }
}

func (c *Crawler) extractData(body io.Reader, baseURL string) (title string, links, emails, phones, ips []string) {
    doc, err := html.Parse(body)
    if err != nil {
        return
    }
    
    emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
    phoneRegex := regexp.MustCompile(`(\+?[0-9]{1,3}[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}`)
    ipRegex := regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
    
    var f func(*html.Node)
    f = func(n *html.Node) {
        if n.Type == html.ElementNode {
            if n.Data == "title" && n.FirstChild != nil {
                title = n.FirstChild.Data
            }
            if n.Data == "a" {
                for _, attr := range n.Attr {
                    if attr.Key == "href" {
                        link := c.resolveURL(attr.Val, baseURL)
                        if strings.Contains(link, c.targetDomain) {
                            links = append(links, link)
                        }
                    }
                }
            }
        }
        if n.Type == html.TextNode {
            emails = append(emails, emailRegex.FindAllString(n.Data, -1)...)
            phones = append(phones, phoneRegex.FindAllString(n.Data, -1)...)
            ips = append(ips, ipRegex.FindAllString(n.Data, -1)...)
        }
        for child := n.FirstChild; child != nil; child = child.NextSibling {
            f(child)
        }
    }
    f(doc)
    
    return
}

func (c *Crawler) resolveURL(href, base string) string {
    parsedBase, err := url.Parse(base)
    if err != nil {
        return href
    }
    parsedHref, err := url.Parse(href)
    if err != nil {
        return href
    }
    resolved := parsedBase.ResolveReference(parsedHref)
    return resolved.String()
}

func (c *Crawler) fetchAndParse(rawURL string) CrawlResult {
    result := CrawlResult{
        URL:       rawURL,
        CrawledAt: time.Now(),
    }
    
    req, err := http.NewRequest("GET", rawURL, nil)
    if err != nil {
        return result
    }
    
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Luxior OSINT Agent")
    req.Header.Set("Accept", "text/html,application/xhtml+xml")
    
    resp, err := c.client.Do(req)
    if err != nil {
        return result
    }
    defer resp.Body.Close()
    
    result.StatusCode = resp.StatusCode
    result.ContentType = resp.Header.Get("Content-Type")
    
    if resp.StatusCode != 200 {
        return result
    }
    
    if strings.Contains(result.ContentType, "text/html") {
        title, links, emails, phones, ips := c.extractData(resp.Body, rawURL)
        result.Title = title
        result.Links = links
        result.Emails = uniqueStrings(emails)
        result.PhoneNumbers = uniqueStrings(phones)
        result.IpAddresses = uniqueStrings(ips)
    }
    
    return result
}

func uniqueStrings(slice []string) []string {
    keys := make(map[string]bool)
    var list []string
    for _, entry := range slice {
        if !keys[entry] {
            keys[entry] = true
            list = append(list, entry)
        }
    }
    return list
}

func (c *Crawler) worker() {
    defer c.wg.Done()
    
    for url := range c.urlQueue {
        if _, loaded := c.visited.LoadOrStore(url, true); loaded {
            continue
        }
        
        result := c.fetchAndParse(url)
        c.results <- result
        
        jsonData, _ := json.Marshal(result)
        c.redis.LPush(c.redis.Context(), "lux:crawl:queue", jsonData)
        
        c.saveToDB(result)
        
        for _, link := range result.Links {
            c.urlQueue <- link
        }
    }
}

func (c *Crawler) saveToDB(result CrawlResult) {
    emails := strings.Join(result.Emails, ",")
    phones := strings.Join(result.PhoneNumbers, ",")
    ips := strings.Join(result.IpAddresses, ",")
    
    stmt, _ := c.db.Prepare(`INSERT OR REPLACE INTO crawled_pages 
        (url, title, status_code, content_type, emails, phones, ips, crawled_at) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
    defer stmt.Close()
    
    stmt.Exec(result.URL, result.Title, result.StatusCode, result.ContentType, 
        emails, phones, ips, result.CrawledAt)
}

func (c *Crawler) Start(startURL string) {
    c.urlQueue <- startURL
    
    for i := 0; i < 20; i++ {
        c.wg.Add(1)
        go c.worker()
    }
    
    go func() {
        c.wg.Wait()
        close(c.results)
    }()
    
    for result := range c.results {
        fmt.Printf("[CRAWL] %s | %d | %s\n", result.URL, result.StatusCode, result.Title)
    }
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: lux_crawler <target_url> [max_depth] [max_urls]")
        return
    }
    
    target := os.Args[1]
    depth := 3
    maxUrls := 1000
    
    if len(os.Args) > 2 {
        fmt.Sscanf(os.Args[2], "%d", &depth)
    }
    if len(os.Args) > 3 {
        fmt.Sscanf(os.Args[3], "%d", &maxUrls)
    }
    
    fmt.Printf("[LUX CRAWLER] Starting crawl on %s (depth: %d, max: %d)\n", target, depth, maxUrls)
    
    crawler := NewCrawler(target, depth, maxUrls)
    crawler.Start(target)
    
    fmt.Printf("[LUX CRAWLER] Crawl completed\n")
}
