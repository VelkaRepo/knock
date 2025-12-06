package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

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

	fmt.Printf("[*] Knocking on %s with %d threads...\n", *targetURL, *threads)

	// 2. Setup Channels (The "Conveyor Belt")
	jobs := make(chan string)
	var wg sync.WaitGroup

	// 3. Define the Client (Shared by all workers)
	client := &http.Client{Timeout: 5 * time.Second}

	// 4. Spawn Workers (The "Minions")
	for i := 0; i < *threads; i++ {
		wg.Add(1)
		
		go func() {
			defer wg.Done()
			
			for url := range jobs {
				resp, err := client.Get(url)
				if err != nil {
					continue
				}
				
				if resp.StatusCode == 200 {
					fmt.Printf("[+] FOUND: %s (Status: 200)\n", url)
				}
				resp.Body.Close()
			}
		}()
	}

	// 5. Feed the Conveyor Belt (Main Thread)
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

	// 6. Close the shop
	close(jobs)
	wg.Wait()
}