# gsc-cli

Query Google Search Console from the terminal. Pull top queries, compare date ranges, export CSVs, and flag ranking drops.

## Installation

```bash
# From source
go install ./cmd/gsc

# Or build locally
make build
./bin/gsc
```

## Setup

### 1. Create Google Cloud OAuth Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Create a new project (or select existing)
3. Enable the **Search Console API**
4. Go to **APIs & Services → Credentials**
5. Click **Create Credentials → OAuth client ID**
6. Select **Desktop app**
7. Download the `client_secret.json` file

### 2. Authenticate

```bash
gsc auth login --client-secret ~/path/to/client_secret.json
```

This opens your browser for Google OAuth consent. Tokens are stored securely in your OS keychain.

## Usage

### Top Queries

```bash
# Last 28 days, top 100 queries
gsc queries

# Last 7 days
gsc queries --days 7

# Custom date range
gsc queries --start 2025-01-01 --end 2025-01-15

# Top 500 queries
gsc queries --limit 500

# Filter by page pattern
gsc queries --filter "page:*/blog/*"

# Group by page instead of query
gsc queries --dimension page

# Export to CSV
gsc queries --csv output.csv

# JSON output
gsc queries --json
```

### Compare Date Ranges

```bash
# This week vs last week
gsc compare --period week

# This month vs last month
gsc compare --period month

# Custom date ranges
gsc compare --from-start 2025-01-01 --from-end 2025-01-15 \
            --to-start 2024-12-15 --to-end 2024-12-31

# Sort by impressions
gsc compare --sort impressions

# Export comparison
gsc compare --csv comparison.csv
```

### Detect Ranking Drops

```bash
# Find queries that dropped >5 positions
gsc drops

# Lower threshold (>3 positions)
gsc drops --threshold 3

# Only queries with significant traffic
gsc drops --min-clicks 10

# Compare 14-day periods
gsc drops --days 14

# Export drops
gsc drops --csv drops.csv
```

### List Sites

```bash
# Show all sites you have access to
gsc sites
```

This displays the exact site URL format to use.

### Configuration

```bash
# Set default site
gsc config set-site sc-domain:example.com

# Show current config
gsc config show
```

### Authentication

```bash
# Check auth status
gsc auth status

# Log out (removes stored token)
gsc auth logout
```

## Global Flags

| Flag | Description |
|------|-------------|
| `-s, --site` | Override default site URL |
| `--json` | Output as JSON |
| `--no-color` | Disable colored output |

## Shell Completion

```bash
# Bash
source <(gsc completion bash)

# Zsh
source <(gsc completion zsh)

# Fish
gsc completion fish | source
```

## Site URL Formats

Google Search Console uses two property types:

- **Domain property**: `sc-domain:example.com`
- **URL prefix**: `https://example.com/`

Use the exact format shown in your Search Console dashboard.

## Security

- OAuth tokens are stored in your OS keychain (macOS Keychain, Windows Credential Manager, or Linux Secret Service)
- The `client_secret.json` path is stored in `~/.config/gsc-cli/config.yaml`
- No credentials are stored in plaintext files

## License

MIT
