package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ANSI Color Codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
)

var soft404Size int64 = -1

func main() {
	// 1. Arguments
	targetURL := flag.String("u", "", "Target URL")
	wordlistPtr := flag.String("w", "wordlist.txt", "Path to wordlist")
	threads := flag.Int("t", 20, "Number of concurrent threads")
	extPtr := flag.String("x", "", "Extensions to check (comma separated, e.g., php,html)")
	flag.Parse()

	if *targetURL == "" {
		// Print Error in RED
		fmt.Printf("%s[-] Error: You must provide a target URL (-u)%s\n", Red, Reset)
		fmt.Println("Usage: knock -u <url> -w <wordlist> -t <threads> -x <extensions>")
		os.Exit(1)
	}

	// 2. Process Extensions
	extensions := []string{""}
	if *extPtr != "" {
		extList := strings.Split(*extPtr, ",")
		for _, ext := range extList {
			extensions = append(extensions, "."+ext)
		}
	}

	client := &http.Client{Timeout: 5 * time.Second}

	// 3. CALIBRATION (Cyan for Info)
	fmt.Printf("%s[*] Calibrating Soft 404 on %s...%s\n", Cyan, *targetURL, Reset)
	calibrate(client, *targetURL)

	fmt.Printf("%s[*] Knocking with %d threads (checking %d variations per word)...%s\n", Yellow, *threads, len(extensions), Reset)

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
					fmt.Printf("%s[+] FOUND: %s%s (Size: %d)\n", Green, Reset, url, length)
				}
			}
		}()
	}

	// 5. Feed the Belt
	file, err := os.Open(*wordlistPtr)
	if err != nil {
		fmt.Printf("%s[-] Error: Could not open wordlist '%s'%s\n", Red, *wordlistPtr, Reset)
		os.Exit(1)
	}
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := scanner.Text()
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
		fmt.Printf("%s[-] Calibration failed: Could not connect to target.%s\n", Red, Reset)
		os.Exit(1)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	soft404Size = int64(len(bodyBytes))
	
	fmt.Printf("%s[*] Soft 404 size established: %d bytes%s\n", Cyan, soft404Size, Reset)
}