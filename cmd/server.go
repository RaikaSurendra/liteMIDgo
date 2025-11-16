package cmd

import (
	"fmt"
	"os"

	"litemidgo/config"
	"litemidgo/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the LiteMIDgo server with dashboard",
	Long: `Start the LiteMIDgo middleware server with a beautiful dashboard UI.
The dashboard shows server status, configuration, statistics, and allows
you to start/stop the server interactively.`,
	Run: func(cmd *cobra.Command, args []string) {
		startServerWithDashboard()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

func startServerWithDashboard() {
	// Load configuration
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		fmt.Printf("❌ Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("❌ Configuration validation failed: %v\n", err)
		os.Exit(1)
	}

	// Create and run the Bubble Tea server dashboard UI
	model := ui.NewServerDashboardModel(cfg)
	program := tea.NewProgram(model, tea.WithAltScreen())

	_, err = program.Run()
	if err != nil {
		fmt.Printf("Error running server dashboard: %v\n", err)
		os.Exit(1)
	}
}
