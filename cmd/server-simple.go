package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"litemidgo/config"
	"litemidgo/internal/server"

	"github.com/spf13/cobra"
)

var serverSimpleCmd = &cobra.Command{
	Use:   "server-simple",
	Short: "Start the LiteMIDgo server in simple mode (no TUI)",
	Long: `Start the LiteMIDgo middleware server in simple mode without the dashboard UI.
This is useful for running the server in the background or in scripts.`,
	Run: func(cmd *cobra.Command, args []string) {
		startServerSimple()
	},
}

func init() {
	rootCmd.AddCommand(serverSimpleCmd)
}

func startServerSimple() {
	// Load configuration
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	log.Printf("Starting LiteMIDgo server on %s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("ServiceNow instance: %s", cfg.ServiceNow.Instance)

	// Create and start server
	srv := server.NewServer(cfg)

	// Setup graceful shutdown
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down LiteMIDgo server...")
	if err := srv.Stop(); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	log.Println("Server stopped")
}
