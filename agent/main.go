package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	serverURL string
	interval  int
	once      bool
	debug     bool
)

type AgentConfig struct {
	ServerURL string `json:"server_url"`
	Interval  int    `json:"interval"`
	AgentName string `json:"agent_name"`
}

type Payload struct {
	Agent   string      `json:"agent"`
	Topic   string      `json:"topic"`
	Name    string      `json:"name"`
	Source  string      `json:"source"`
	Payload interface{} `json:"payload"`
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "litemidgo-agent",
		Short: "LiteMIDgo SensuGo-compatible agent",
		Long: `A lightweight monitoring agent that collects system metrics
and sends them to LiteMIDgo server for ServiceNow integration.`,
	}

	var collectCmd = &cobra.Command{
		Use:   "collect",
		Short: "Collect and display system metrics",
		Run: func(cmd *cobra.Command, args []string) {
			PrintMetricsJSON()
		},
	}

	var sendCmd = &cobra.Command{
		Use:   "send",
		Short: "Collect and send metrics to LiteMIDgo server",
		Run: func(cmd *cobra.Command, args []string) {
			sendMetrics()
		},
	}

	var daemonCmd = &cobra.Command{
		Use:   "daemon",
		Short: "Run as daemon, sending metrics periodically",
		Run: func(cmd *cobra.Command, args []string) {
			runDaemon()
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&serverURL, "server", "s", "http://localhost:8080", "LiteMIDgo server URL")
	rootCmd.PersistentFlags().IntVarP(&interval, "interval", "i", 60, "Collection interval in seconds")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Show JSON payload being sent")
	daemonCmd.Flags().BoolVar(&once, "once", false, "Send metrics once and exit")

	rootCmd.AddCommand(collectCmd)
	rootCmd.AddCommand(sendCmd)
	rootCmd.AddCommand(daemonCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func sendMetrics() {
	metrics, err := CollectSystemMetrics()
	if err != nil {
		log.Fatalf("Failed to collect metrics: %v", err)
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	payload := Payload{
		Agent:  "litemidgo-agent",
		Topic:  "endpointData",
		Name:   hostname,
		Source: hostname,
		Payload: map[string]interface{}{
			"endpoint_metrics": map[string]interface{}{
				"hostname":         metrics.Hostname,
				"collection_time":  time.Now().UTC().Format(time.RFC3339),
				"agent_version":    "1.0.0",
				"operating_system": metrics.OS,
				"cpu_metrics":      metrics.CPU,
				"memory_metrics":   metrics.Memory,
				"disk_metrics":     metrics.Disk,
				"network_metrics":  metrics.Network,
				"runtime_metrics":  metrics.Runtime,
				"raw_timestamp":    metrics.Timestamp,
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Failed to marshal payload: %v", err)
	}

	apiURL := fmt.Sprintf("%s/proxy/ecc_queue", serverURL)

	// Debug output - show formatted JSON
	if debug {
		prettyJSON, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			fmt.Printf("Error formatting JSON: %v\n", err)
		} else {
			fmt.Printf("🔍 Debug - JSON payload being sent:\n")
			fmt.Printf("%s\n", string(prettyJSON))
			fmt.Printf("📡 Sending to: %s\n\n", apiURL)
		}
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to send metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Server returned status: %d", resp.StatusCode)
	}

	fmt.Printf("✅ Metrics sent successfully to %s\n", serverURL)
}

func runDaemon() {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	fmt.Printf("🚀 Starting LiteMIDgo agent for %s\n", hostname)
	fmt.Printf("📡 Sending metrics to %s every %d seconds\n", serverURL, interval)
	fmt.Printf("🔄 Press Ctrl+C to stop\n\n")

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	// Send initial metrics
	sendMetrics()

	for range ticker.C {
		sendMetrics()
	}
}
