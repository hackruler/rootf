package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type CertEntry struct {
	CommonName string `json:"common_name"`
	NameValue  string `json:"name_value"`
}

var (
	inputFile string
	helpFlag  bool
)

func init() {
	flag.StringVar(&inputFile, "l", "", "Specify the domain file. If not provided, input can be piped.")
	flag.BoolVar(&helpFlag, "h", false, "Show help")
}

func showHelp() {
	fmt.Println("   Usage: go run rootf.go -l <root domains file>\n")
	fmt.Println("   Options:")
	fmt.Println("   -h                              Help Menu")
	fmt.Println("   -l <input file name>            Specify the domain file. If not provided, input can be piped.")
}

func readInput() ([]string, error) {
	var reader io.Reader
	if inputFile != "" {
		file, err := os.Open(inputFile)
		if err != nil {
			return nil, fmt.Errorf("error: cannot open input file: %v", err)
		}
		defer file.Close()
		reader = file
	} else {
		stat, err := os.Stdin.Stat()
		if err != nil {
			return nil, fmt.Errorf("error: failed to check stdin: %v", err)
		}
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return nil, fmt.Errorf("error: no input file or piped data")
		}
		reader = os.Stdin
	}

	var domains []string
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			domains = append(domains, line)
		}
	}
	return domains, nil
}

func fetchFromCRT(domain string) ([]string, error) {
	url := fmt.Sprintf("https://crt.sh/?q=%s&output=json", domain)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var entries []CertEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}

	var subdomains []string
	for _, entry := range entries {
		if strings.HasPrefix(entry.CommonName, "*.") {
			subdomains = append(subdomains, strings.TrimPrefix(entry.CommonName, "*.")) // remove *
		}
		for _, name := range strings.Split(entry.NameValue, "\n") {
			if strings.HasPrefix(name, "*.") {
				subdomains = append(subdomains, strings.TrimPrefix(name, "*.")) // remove *
			}
		}
	}
	return subdomains, nil
}

func processDomain(domain string) []string {
	subdomains, err := fetchFromCRT(domain)
	if err != nil || len(subdomains) == 0 {
		// Retry once after 20s
		fmt.Fprintln(os.Stderr, "[!] No response for", domain, "- retrying after 20s...")
		time.Sleep(20 * time.Second)
		subdomains, err = fetchFromCRT(domain)
		if err != nil || len(subdomains) == 0 {
			fmt.Fprintln(os.Stderr, "[!] Still no response for", domain, "- skipping.")
			return nil
		}
	}

	filtered := []string{}
	for _, sub := range subdomains {
		if strings.HasSuffix(sub, "."+domain) || sub == domain {
			filtered = append(filtered, sub)
		}
	}
	return filtered
}

func main() {
	flag.Parse()

	if helpFlag {
		showHelp()
		return
	}

	domains, err := readInput()
	if err != nil {
		fmt.Fprintln(os.Stderr, "[!] "+err.Error())
		showHelp()
		os.Exit(1)
	}

	allResults := make(map[string]struct{})
	for _, domain := range domains {
		domain = strings.TrimSpace(domain)
		if domain == "" {
			continue
		}
		results := processDomain(domain)
		for _, r := range results {
			allResults[r] = struct{}{}
		}
		time.Sleep(1 * time.Second)
	}

	// Sort and print clean output to stdout
	final := []string{}
	for sub := range allResults {
		final = append(final, sub)
	}
	sort.Strings(final)

	for _, sub := range final {
		fmt.Println(sub)
	}
}

