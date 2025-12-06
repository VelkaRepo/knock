package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// Global variable to store the size of the "fake" 404 page
var soft404Size int64 = -1

func main() {
	// 1. Arguments
	targetURL := flag.String("u", "", "Target URL")
	wordlistPtr := flag.String("w", "wordlist.txt", "Path to wordlist")
	threads := flag.Int("t", 20, "Number of concurrent threads")
	flag.Parse()

	if *targetURL == "" {
		fmt.Println("Usage: knock -u <url> -w <wordlist> -t <threads>")
		os.Exit(1)
	}

	// 2. Setup Client
	client := &http.Client{Timeout: 5 * time.Second}

	// 3. CALIBRATION PHASE
	fmt.Printf("[*] Calibrating Soft 404 detection on %s...\n", *targetURL)
	calibrate(client, *targetURL)

	fmt.Printf("[*] Knocking with %d threads...\n", *threads)

	// 4. Setup Conveyor Belt
	jobs := make(chan string)
	var wg sync.WaitGroup

	// 5. Spawn Workers
	for i := 0; i < *threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range jobs {
				resp, err := client.Get(url)
				if err != nil {
					continue
				}
				
				// READ the body to get the true length 
				// (ContentLength header isn't always accurate, so we discard body but count bytes)
				bodyBytes, _ := io.ReadAll(resp.Body)
				length := int64(len(bodyBytes))
				resp.Body.Close()

				// THE SMART CHECK
				if resp.StatusCode == 200 {
					// If the page size matches the "Not Found" page size, ignore it!
					if length == soft404Size {
						continue 
					}
					
					fmt.Printf("[+] FOUND: %s (Status: 200, Size: %d)\n", url, length)
				}
			}
		}()
	}

	// 6. Feed the Conveyor Belt
	file, err := os.Open(*wordlistPtr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := scanner.Text()
		fullURL := fmt.Sprintf("%s/%s", *targetURL, word)
		jobs <- fullURL 
	}
	file.Close()

	close(jobs)
	wg.Wait()
}

// calibrate sends a request to a nonsense URL to see what a 404 looks like
func calibrate(client *http.Client, baseURL string) {
	// Request a random junk URL
	junkURL := baseURL + "/" + "thisshouldnotexist_random_12345"
	resp, err := client.Get(junkURL)
	if err != nil {
		fmt.Println("[-] Calibration failed: Could not connect.")
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Measure the length of the error page
	bodyBytes, _ := io.ReadAll(resp.Body)
	soft404Size = int64(len(bodyBytes))

	fmt.Printf("[*] Calibration complete. Soft 404 pages are approx %d bytes.\n", soft404Size)
}