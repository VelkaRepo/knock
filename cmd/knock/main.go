package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings" // NEW: Needed to split the commas
	"sync"
	"time"
)

var soft404Size int64 = -1

func main() {
	// 1. Arguments
	targetURL := flag.String("u", "", "Target URL")
	wordlistPtr := flag.String("w", "wordlist.txt", "Path to wordlist")
	threads := flag.Int("t", 20, "Number of concurrent threads")
	extPtr := flag.String("x", "", "Extensions to check (comma separated, e.g., php,html)") // NEW FLAG
	flag.Parse()

	if *targetURL == "" {
		fmt.Println("Usage: knock -u <url> -w <wordlist> -t <threads> -x <extensions>")
		os.Exit(1)
	}

	// 2. Process Extensions
	// If user types "php,html", we create a list: ["", ".php", ".html"]
	// We add "" (empty string) so we always check the original word too!
	extensions := []string{""}
	if *extPtr != "" {
		// Split "php,html" into ["php", "html"]
		extList := strings.Split(*extPtr, ",")
		for _, ext := range extList {
			// Add the dot, e.g., ".php"
			extensions = append(extensions, "."+ext)
		}
	}

	client := &http.Client{Timeout: 5 * time.Second}

	// 3. CALIBRATION
	fmt.Printf("[*] Calibrating Soft 404 on %s...\n", *targetURL)
	calibrate(client, *targetURL)

	fmt.Printf("[*] Knocking with %d threads (checking %d variations per word)...\n", *threads, len(extensions))

	jobs := make(chan string)
	var wg sync.WaitGroup

	// 4. Workers
	for i := 0; i < *threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range jobs {
				resp, err := client.Get(url)
				if err != nil {
					continue
				}
				
				bodyBytes, _ := io.ReadAll(resp.Body)
				length := int64(len(bodyBytes))
				resp.Body.Close()

				if resp.StatusCode == 200 {
					if length == soft404Size {
						continue 
					}
					// Found!
					fmt.Printf("[+] FOUND: %s (Status: 200, Size: %d)\n", url, length)
				}
			}
		}()
	}

	// 5. Feed the Belt
	file, err := os.Open(*wordlistPtr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := scanner.Text()
		
		// NEW LOOP: For every word, generate all extension variations
		for _, ext := range extensions {
			fullURL := fmt.Sprintf("%s/%s%s", *targetURL, word, ext)
			jobs <- fullURL
		}
	}
	file.Close()

	close(jobs)
	wg.Wait()
}

func calibrate(client *http.Client, baseURL string) {
	junkURL := baseURL + "/" + "thisshouldnotexist_random_12345"
	resp, err := client.Get(junkURL)
	if err != nil {
		fmt.Println("[-] Calibration failed: Could not connect.")
		os.Exit(1)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	soft404Size = int64(len(bodyBytes))
}