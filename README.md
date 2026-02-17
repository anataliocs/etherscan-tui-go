# Ethereum Transaction Explorer

A modern, terminal-based Ethereum transaction explorer built with Go and the [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework. This tool allows you to quickly fetch and display details for any Ethereum transaction hash using the Etherscan API V2.

## Features

Built with `bubbletea`, `bubbles`, and `lipgloss`.

## Prerequisites

- [Go](https://go.dev/doc/install) 1.26 or later.
- An [Etherscan API Key](https://etherscan.io/apis).

## Setup

Create a `.env` file in the project root (or update the existing one):
   ```text
   ETHERSCAN_API_KEY=your_etherscan_api_key_here
   ```

**Install dependencies**:
   ```bash
   go mod tidy
   ```

## Installation & Running

### Build from source
```bash
go build -o ethereum-explorer ./cmd/ethereum-explorer
./ethereum-explorer
```

### Run directly
```bash
go run ./cmd/ethereum-explorer
```

## Project Structure

- `cmd/ethereum-explorer/`: Application entry point.
- `internal/etherscan/`: Client for interacting with the Etherscan API V2.
- `internal/ui/`: Bubble Tea models, views, and Lipgloss styles.
- `internal/config/`: Configuration and environment variable management.
- `.env`: Local environment variables (ignored by git).
- `main.go`: Legacy entry point (deprecated).

## License

[MIT](LICENSE)
