# Ethereum Transaction Explorer

A terminal(TUI) Ethereum transaction explorer built with Go and
the [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework. Fetch, display and explore details for any Ethereum transaction hash 
using the Etherscan API V2 all in your terminal.

Built with `bubbletea`, `bubbles`, and `lipgloss`.

### Current Supported EVM Networks
- [Ethereum](https://etherscan.io/)
- [Sepolia](https://sepolia.etherscan.io/)

## Prerequisites

- [Go](https://go.dev/doc/install) 1.26 or later.
- An [Etherscan API Key](https://etherscan.io/apis).

## Setup

Create `.env` file in project root:

```shell
cp .env.example .env
```

Add your Etherscan API key to the `.env` file:

```text
ETHERSCAN_API_KEY=your_etherscan_api_key_here
```

**Install dependencies**:

```bash
go mod tidy
```

### Build

```bash
go build -o ethereum-explorer ./cmd/ethereum-explorer
```

Run the binary:

```shell
./ethereum-explorer
```

### Run directly

```bash
go run ./cmd/ethereum-explorer
```

## Tests

Run Tests
```bash
go test ./... -v
```

Skip cache
```bash
go test -count=1 ./... -v
```

Run tests for a specific package (e.g., `etherscan`):

```bash
go test ./internal/etherscan/...
```

## Project Structure

- `cmd/ethereum-explorer/`: Application entry point.
- `internal/etherscan/`: Client for interacting with the Etherscan API V2.
    - `client.go`: Main client and API request logic.
    - `types.go`: Struct definitions for Etherscan responses and the internal `Transaction` type.
    - `json.go`: JSON unmarshaling and response extraction helpers.
    - `retry.go`: HTTP request implementation with exponential backoff.
    - `format.go`: Formatting utilities for ETH values, gas prices, and transaction types.
    - `convert.go`: Conversion helpers (hex-to-decimal, confirmations calculation, etc.).
- `internal/model/`: Main Bubble Tea application model and state management.
    - `model.go`: TUI state, initialization, and sub-component orchestration.
    - `update.go`: Message handling and state transitions.
    - `view.go`: Main UI rendering logic delegating to components.
- `internal/tui/`: TUI-specific components and styling following the MVU pattern.
    - `components/`: Reusable UI elements (header, footer, input, loader, transaction, errorview).
    - `context/`: Shared `ProgramContext` for global state like terminal dimensions and theme.
    - `theme/`: Centralized styles and adaptive color definitions using Lipgloss.
- `internal/config/`: Configuration and environment variable management.
- `.env`: Local environment variables (ignored by git).
- `main.go`: Deprecated entry point.

## Maintained by:

[Hella Web3 Labs](https://hella.website/)

## License

[MIT](LICENSE)
