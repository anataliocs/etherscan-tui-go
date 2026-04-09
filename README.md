# Ethereum Transaction Explorer

A modern, terminal-based Ethereum transaction explorer built with Go and
the [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework. This tool allows you to quickly fetch and
display details for any Ethereum transaction hash using the Etherscan API V2.

Built with `bubbletea`, `bubbles`, and `lipgloss`.

## Prerequisites

- [Go](https://go.dev/doc/install) 1.26 or later.
- An [Etherscan API Key](https://etherscan.io/apis).

## Setup

Create a `.env` file in the project root:

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

## Installation & Running

### Build from source

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

## Running Tests

To run the unit tests for the project:

```bash
go test ./... -v
```

Skip cache
```bash
go test -count=1 ./... -v
```

To run tests for a specific package (e.g., `etherscan`):

```bash
go test ./internal/etherscan/...
```

**./cmd/ethereum-explorer/main.go**
```go
package main

import (
	"fmt"
	"os"

	"awesomeProject/internal/config"
	"awesomeProject/internal/etherscan"
	"awesomeProject/internal/model"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	config.LoadEnv()

	apiKey := config.APIKey()
	if apiKey == "" {
		fmt.Println("Error: ETHERSCAN_API_KEY environment variable is not set.")
		fmt.Println("Please create a .env file with your Etherscan API key.")
		os.Exit(1)
	}

	client := etherscan.NewClient(apiKey)
	m := model.New(client)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
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

## Documentation

All packages and functions include comprehensive Go documentation comments, explaining parameters and return values.

## License

[MIT](LICENSE)
