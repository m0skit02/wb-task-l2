package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type Crawler struct {
	baseURL   *url.URL
	maxDepth  int
	visited   map[string]bool
	mu        sync.Mutex
	client    *http.Client
	outputDir string
	sem       chan struct{}
}

func NewCrawler(start string, depth int) (*Crawler, error) {
	u, err := url.Parse(start)
	if err != nil {
		return nil, err
	}

	return &Crawler{
		baseURL:  u,
		maxDepth: depth,
		visited:  make(map[string]bool),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		outputDir: "mirror",
		sem:       make(chan struct{}, 10), // ограничение параллелизма
	}, nil
}

func (c *Crawler) Start() error {
	return c.crawl(c.baseURL.String(), 0)
}

func (c *Crawler) crawl(rawURL string, depth int) error {
	if depth > c.maxDepth {
		return nil
	}

	c.mu.Lock()
	if c.visited[rawURL] {
		c.mu.Unlock()
		return nil
	}
	c.visited[rawURL] = true
	c.mu.Unlock()

	c.sem <- struct{}{}
	defer func() { <-c.sem }()

	fmt.Println("Downloading:", rawURL)

	resp, err := c.client.Get(rawURL)
	if err != nil {
		fmt.Println("request error:", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("bad status:", resp.Status)
		return nil
	}

	contentType := resp.Header.Get("Content-Type")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("read error:", err)
		return nil
	}

	localPath := c.saveFile(rawURL, body)

	// Обрабатываем только HTML
	if strings.Contains(contentType, "text/html") {
		doc, err := html.Parse(bytes.NewReader(body))
		if err != nil {
			return nil
		}

		links := extractLinks(doc, c.baseURL)

		// переписываем ссылки
		rewriteLinks(doc, c.baseURL)

		// сохраняем уже изменённый HTML
		var buf bytes.Buffer
		html.Render(&buf, doc)
		os.WriteFile(localPath, buf.Bytes(), 0644)

		var wg sync.WaitGroup

		for _, link := range links {
			wg.Add(1)
			go func(l string) {
				defer wg.Done()
				c.crawl(l, depth+1)
			}(link)
		}

		wg.Wait()
	}

	return nil
}

func (c *Crawler) saveFile(rawURL string, data []byte) string {
	u, _ := url.Parse(rawURL)

	path := u.Path

	if path == "" || strings.HasSuffix(path, "/") {
		path = path + "index.html"
	}

	localPath := filepath.Join(c.outputDir, u.Host, path)

	err := os.MkdirAll(filepath.Dir(localPath), os.ModePerm)
	if err != nil {
		fmt.Println("mkdir error:", err)
	}

	err = os.WriteFile(localPath, data, 0644)
	if err != nil {
		fmt.Println("write error:", err)
	}

	return localPath
}

func extractLinks(n *html.Node, base *url.URL) []string {
	var links []string

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, attr := range n.Attr {
				if attr.Key == "href" || attr.Key == "src" {
					u, err := url.Parse(attr.Val)
					if err != nil {
						continue
					}

					abs := base.ResolveReference(u)

					if abs.Host == base.Host {
						links = append(links, abs.String())
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(n)
	return unique(links)
}

func rewriteLinks(n *html.Node, base *url.URL) {
	var f func(*html.Node)

	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for i, attr := range n.Attr {
				if attr.Key == "href" || attr.Key == "src" {
					u, err := url.Parse(attr.Val)
					if err != nil {
						continue
					}

					abs := base.ResolveReference(u)

					if abs.Host == base.Host {
						local := localPathFromURL(abs)
						n.Attr[i].Val = local
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(n)
}

func localPathFromURL(u *url.URL) string {
	path := u.Path

	if path == "" || strings.HasSuffix(path, "/") {
		path += "index.html"
	}

	return "." + path
}

func unique(input []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, v := range input {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	return result
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <url> [depth]")
		return
	}

	startURL := os.Args[1]
	depth := 2

	if len(os.Args) >= 3 {
		fmt.Sscanf(os.Args[2], "%d", &depth)
	}

	crawler, err := NewCrawler(startURL, depth)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	err = crawler.Start()
	if err != nil {
		fmt.Println("crawl error:", err)
	}
}
