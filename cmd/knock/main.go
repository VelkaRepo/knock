package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	// 1. Define CLI Arguments (Flags)
	// usage: -u <url> -w <wordlist>
	targetURL := flag.String("u", "", "Target URL (e.g., http://localhost:8000)")
	wordlistPtr := flag.String("w", "wordlist.txt", "Path to wordlist")
	
	flag.Parse() // This actually processes the command line args

	// 2. Validation: Did the user provide a URL?
	if *targetURL == "" {
		fmt.Println("[-] Error: You must provide a target URL with -u")
		fmt.Println("Usage: knock -u <url> -w <wordlist>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("[*] Knocking on %s using %s...\n", *targetURL, *wordlistPtr)

	// 3. Open Wordlist
	file, err := os.Open(*wordlistPtr)
	if err != nil {
		fmt.Printf("[-] Error: Could not open wordlist '%s'\n", *wordlistPtr)
		os.Exit(1)
	}
	defer file.Close()

	// 4. Setup Client
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 5. Scan Loop
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := scanner.Text()
		// Handle slash logic: ensure there is exactly one / between url and word
		// Simple approach for now:
		fullURL := fmt.Sprintf("%s/%s", *targetURL, word)

		resp, err := client.Get(fullURL)
		if err != nil {
			continue
		}

		if resp.StatusCode == 200 {
			fmt.Printf("[+] FOUND: /%s (Status: 200)\n", word)
		}
		
		resp.Body.Close()
	}
}