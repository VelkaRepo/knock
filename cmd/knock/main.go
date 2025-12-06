package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
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
	targetURL := flag.String("u", "", "Target URL")
	wordlistPtr := flag.String("w", "wordlist.txt", "Path to wordlist")
	threads := flag.Int("t", 20, "Number of concurrent threads")
	extPtr := flag.String("x", "", "Extensions (comma separated)")
	flag.Parse()

	if *targetURL == "" {
		fmt.Printf("%s[-] Error: Target URL required (-u)%s\n", Red, Reset)
		os.Exit(1)
	}

	// 1. Prepare Extensions
	extensions := []string{""}
	if *extPtr != "" {
		for _, ext := range strings.Split(*extPtr, ",") {
			extensions = append(extensions, "."+ext)
		}
	}

	// 2. Count Total Work (Lines * Extensions)
	totalLines, err := countLines(*wordlistPtr)
	if err != nil {
		fmt.Printf("%s[-] Error reading wordlist: %v%s\n", Red, err, Reset)
		os.Exit(1)
	}
	totalRequests := uint64(totalLines * len(extensions))

	client := &http.Client{Timeout: 5 * time.Second}

	// 3. Calibration
	fmt.Printf("%s[*] Calibrating Soft 404 on %s...%s\n", Cyan, *targetURL, Reset)
	calibrate(client, *targetURL)

	fmt.Printf("%s[*] Scanning %d targets with %d threads...%s\n", Yellow, totalRequests, *threads, Reset)

	jobs := make(chan string)
	var wg sync.WaitGroup
	var requestsDone uint64 // Atomic counter

	// 4. Progress Monitor (Background Goroutine)
	// Updates the progress bar every 100ms
	go func() {
		for {
			done := atomic.LoadUint64(&requestsDone)
			if done >= totalRequests {
				break
			}
			printProgress(done, totalRequests)
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// 5. Workers
	for i := 0; i < *threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range jobs {
				atomic.AddUint64(&requestsDone, 1)

				resp, err := client.Get(url)
				if err != nil {
					continue
				}

				body, _ := io.ReadAll(resp.Body)
				length := int64(len(body))
				resp.Body.Close()

				if resp.StatusCode == 200 {
					if length == soft404Size {
						continue
					}
					// Clear progress line before printing result
					fmt.Printf("\r\033[K") 
					fmt.Printf("%s[+] FOUND: %s%s (Size: %d)\n", Green, Reset, url, length)
				}
			}
		}()
	}

	// 6. Job Dispatcher
	file, _ := os.Open(*wordlistPtr)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := scanner.Text()
		for _, ext := range extensions {
			jobs <- fmt.Sprintf("%s/%s%s", *targetURL, word, ext)
		}
	}
	file.Close()

	close(jobs)
	wg.Wait()
	
	// Final 100% bar
	printProgress(totalRequests, totalRequests)
	fmt.Println() // New line at end
}

// countLines counts lines in a file efficiently
func countLines(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	count := 0
	buf := make([]byte, 32*1024)
	lineSep := []byte{'\n'}

	for {
		c, err := file.Read(buf)
		count += bytes.Count(buf[:c], lineSep)
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, err
		}
	}
	return count, nil
}

// calibrate determines the size of the 404 page
func calibrate(client *http.Client, baseURL string) {
	junkURL := baseURL + "/" + "R4nd0m_P4th_Check_123"
	resp, err := client.Get(junkURL)
	if err != nil {
		fmt.Printf("%s[-] Calibration failed%s\n", Red, Reset)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	soft404Size = int64(len(body))
}

// printProgress draws the bar: [====>      ] 45%
func printProgress(current, total uint64) {
	percent := float64(current) / float64(total) * 100
	width := 40
	completed := int(float64(width) * (float64(current) / float64(total)))

	bar := strings.Repeat("=", completed)
	if completed < width {
		bar += ">"
		bar += strings.Repeat(" ", width-completed-1)
	} else {
		bar = strings.Repeat("=", width)
	}

	// \r returns cursor to start of line
	fmt.Printf("\r[%s] %.1f%% (%d/%d)", bar, percent, current, total)
}