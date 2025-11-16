package cmd

import (
	"fmt"
	"os"
	"strings"

	"litemidgo/config"
	"litemidgo/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure ServiceNow instance settings",
	Long: `Interactive configuration setup for ServiceNow instance connection.
This will guide you through setting up the required parameters using a beautiful TUI.

Note: You can also set credentials using environment variables:
- SERVICENOW_INSTANCE: Your ServiceNow instance URL
- SERVICENOW_USERNAME: Your ServiceNow username  
- SERVICENOW_PASSWORD: Your ServiceNow password

Environment variables take precedence over config file settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		runConfigSetup()
	},
}

var configTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test connection to configured ServiceNow instance",
	Long: `Test the connection to the configured ServiceNow instance to verify
that the credentials and network connectivity are working correctly.`,
	Run: func(cmd *cobra.Command, args []string) {
		testConnection()
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configTestCmd)
}

func runConfigSetup() {
	// Create and run the Bubble Tea configuration UI
	model := ui.NewConfigModel()
	program := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := program.Run()
	if err != nil {
		fmt.Printf("Error running configuration UI: %v\n", err)
		os.Exit(1)
	}

	// Get the final answers
	configModel := finalModel.(ui.ConfigModel)
	answers := configModel.Answers

	if len(answers) == 0 {
		fmt.Println("Configuration cancelled.")
		return
	}

	// Create config directory if it doesn't exist
	configDir := "./config"
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		os.Mkdir(configDir, 0755)
	}

	// Write configuration to file
	configFile := configDir + "/config.yaml"
	file, err := os.Create(configFile)
	if err != nil {
		fmt.Printf("❌ Failed to create config file: %v\n", err)
		return
	}
	defer file.Close()

	// Parse boolean values
	useHTTPS := strings.ToLower(answers["use_https"]) == "y" || strings.ToLower(answers["use_https"]) == "yes"

	configContent := fmt.Sprintf(`server:
  host: "%s"
  port: %s

servicenow:
  instance: "%s"
  username: "%s"
  password: "%s"
  use_https: %t
  timeout: %s
`, answers["host"], answers["port"], answers["instance"], answers["username"], answers["password"], useHTTPS, answers["timeout"])

	if _, err := file.WriteString(configContent); err != nil {
		fmt.Printf("❌ Failed to write config file: %v\n", err)
		return
	}

	fmt.Printf("✅ Configuration saved to %s\n", configFile)
	fmt.Println()
	fmt.Println("You can now start the server with:")
	fmt.Println("  litemidgo server")
	fmt.Println()
	fmt.Println("Or test the connection with:")
	fmt.Println("  litemidgo config test")
}

func testConnection() {
	// Load configuration to get instance name
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		fmt.Printf("❌ Failed to load configuration: %v\n", err)
		return
	}

	// Create and run the Bubble Tea connection test UI
	model := ui.NewConnectionTestModel(cfg.ServiceNow.Instance)
	program := tea.NewProgram(model, tea.WithAltScreen())

	_, err = program.Run()
	if err != nil {
		fmt.Printf("Error running connection test UI: %v\n", err)
		os.Exit(1)
	}
}
