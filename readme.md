```markdown
# Kraken Portfolio Tracker

A real-time cryptocurrency portfolio tracker for Kraken exchange using WebSocket API.

## Features

- Real-time price updates via WebSocket
- Color-coded price changes (green for increase, red for decrease)
- Sorted display by asset value
- Responsive terminal UI
- Total portfolio value in USD
- Support for multiple cryptocurrencies
- Secure API key management

## Prerequisites

- Go 1.21 or higher
- Kraken API credentials
  - Generate from: https://www.kraken.com/u/security/api
  - Required permissions: Query Funds & Read-only WebSocket access

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/kraken-portfolio.git
cd kraken-portfolio
```

2. Install dependencies:
```bash
make deps
```

3. Create a `.env` file:
```bash
cp .env.example .env
```

4. Add your Kraken API credentials to `.env`:
```env
KRAKEN_API_KEY=your_api_key_here
KRAKEN_API_SECRET=your_api_secret_here
```

## Usage

### Build and Run

Build the project:
```bash
make build
```

Run the application:
```bash
make run
```

### Run Tests

Run all tests:
```bash
make test
```

### Clean

Remove build artifacts:
```bash
make clean
```

## Project Structure

```
kraken-portfolio/
├── cmd/
│   └── main.go         # Application entry point
├── internal/
│   ├── api/           # Kraken API client
│   ├── config/        # Configuration management
│   ├── models/        # Data models
│   └── ui/            # Terminal UI
├── pkg/
│   └── utils/         # Shared utilities
├── test/              # Test files
├── .env.example       # Environment variables template
├── .gitignore        # Git ignore rules
├── go.mod            # Go module definition
├── go.sum            # Go module checksums
└── Makefile          # Build automation
```

## Configuration

The application can be configured using environment variables or a `.env` file:

| Variable | Description | Required |
|----------|-------------|----------|
| KRAKEN_API_KEY | Your Kraken API key | Yes |
| KRAKEN_API_SECRET | Your Kraken API secret | Yes |

## UI Layout

```
╔══════════════════════ KRAKEN PORTFOLIO ══════════════════════╗
║ ASSET     BALANCE          PRICE         VALUE (USD) ║
╠═══════════════════════════════════════════════════════════════╣
║ ETH      1.50000000     $3000.00         4500.00 ║
║ SOL      10.0000000      $100.00         1000.00 ║
╟───────────────────────────────────────────────────────────────╢
║ USD      1000.00             -           1000.00 ║
╠═══════════════════════════════════════════════════════════════╣
║ TOTAL VALUE: $6500.00                                       ║
╚═══════════════════════════════════════════════════════════════╝
```