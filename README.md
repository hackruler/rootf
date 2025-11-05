# rootf

A Go tool for discovering subdomains from Certificate Transparency (CT) logs using crt.sh.

## Description

`rootf` queries the crt.sh Certificate Transparency database to find subdomains for given root domains. It extracts subdomain information from SSL certificate entries and filters them to only include valid subdomains of the specified root domains.

## Features

- Query crt.sh CT logs for subdomain discovery
- Accept input from file or stdin (piping)
- Automatic retry mechanism for failed requests
- Deduplication of results
- Sorted output
- Rate limiting to avoid overwhelming the API

## Installation

### Install using go install (Recommended)

Install directly from GitHub:

```bash
go install github.com/hackruler/rootf@latest
```

### Install from source

Clone the repository and build:

```bash
git clone https://github.com/hackruler/rootf.git
cd rootf
go build -o rootf rootf.go
```

Or build directly:

```bash
go build -o rootf rootf.go
```

### Run directly without building

```bash
go run rootf.go [options]
```

**Note:** Make sure you have Go 1.16 or higher installed for `go install` to work properly.

## Usage

### Using a file

```bash
rootf -l domains.txt
```

### Using stdin (piping)

```bash
echo "example.com" | rootf
```

Or with multiple domains:

```bash
cat domains.txt | rootf
```

### Options

- `-h`: Show help menu
- `-l <file>`: Specify the domain file. If not provided, input can be piped from stdin.

## Examples

### Single domain from file

```bash
echo "example.com" > domains.txt
rootf -l domains.txt
```

### Multiple domains from file

```bash
cat > domains.txt << EOF
example.com
github.com
google.com
EOF

rootf -l domains.txt
```

### Piping input

```bash
echo -e "example.com\ngithub.com" | rootf
```

## Output

The tool outputs unique, sorted subdomains to stdout, one per line:

```
subdomain1.example.com
subdomain2.example.com
subdomain3.example.com
```

## Error Handling

- If a request fails or returns no results, the tool will retry once after a 20-second delay
- If the retry also fails, a warning is printed to stderr and the domain is skipped
- The tool includes a 1-second delay between domain queries to avoid rate limiting

## Requirements

- Go 1.11 or higher
- Internet connection (to query crt.sh API)

## License

This tool is provided as-is for educational and security research purposes.

