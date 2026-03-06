package main

import (
	"fmt"
	"os"

	"awesomeProject/internal/config"
	"awesomeProject/internal/etherscan"
	"awesomeProject/internal/ui"

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
	m := ui.New(client)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
