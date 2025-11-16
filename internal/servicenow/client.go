package servicenow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"litemidgo/config"
)

type Client struct {
	instance   string
	username   string
	password   string
	useHTTPS   bool
	timeout    time.Duration
	httpClient *http.Client
}

type ECCQueuePayload struct {
	Agent    string      `json:"agent"`
	Topic    string      `json:"topic"`
	Name     string      `json:"name"`
	Source   string      `json:"source"`
	Payload  interface{} `json:"payload"`
}

type ECCQueueResponse struct {
	Result struct {
		Agent           string `json:"agent"`
		Signature       string `json:"signature"`
		ResponseTo      string `json:"response_to"`
		SysModCount     string `json:"sys_mod_count"`
		FromSysID       string `json:"from_sys_id"`
		Source          string `json:"source"`
		SysUpdatedOn    string `json:"sys_updated_on"`
		AgentCorrelator string `json:"agent_correlator"`
		Priority        string `json:"priority"`
		SysDomainPath   string `json:"sys_domain_path"`
		ErrorString     string `json:"error_string"`
		Processed       string `json:"processed"`
		Sequence        string `json:"sequence"`
		SysID           string `json:"sys_id"`
		SysUpdatedBy    string `json:"sys_updated_by"`
		FromHost        string `json:"from_host"`
		Payload         string `json:"payload"`
		SysCreatedOn    string `json:"sys_created_on"`
		SysDomain       struct {
			Link  string `json:"link"`
			Value string `json:"value"`
		} `json:"sys_domain"`
		Name        string `json:"name"`
		Topic       string `json:"topic"`
		State       string `json:"state"`
		Queue       string `json:"queue"`
		SysCreatedBy string `json:"sys_created_by"`
	} `json:"result"`
	Error struct {
		Message string `json:"message"`
		Detail  string `json:"detail"`
	} `json:"error"`
}

func NewClient(cfg *config.ServiceNowConfig) *Client {
	return &Client{
		instance: cfg.Instance,
		username: cfg.Username,
		password: cfg.Password,
		useHTTPS: cfg.UseHTTPS,
		timeout:  time.Duration(cfg.Timeout) * time.Second,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

func (c *Client) SendToECCQueue(payload *ECCQueuePayload) (*ECCQueueResponse, error) {
	// Build the URL
	apiURL := fmt.Sprintf("%s://%s/api/now/table/ecc_queue", c.getProtocol(), c.instance)
	
	// Marshal the payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.username, c.password)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code - ServiceNow might return 200 instead of 201
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("ServiceNow API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var eccResp ECCQueueResponse
	if err := json.Unmarshal(body, &eccResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w\nResponse body: %s", err, string(body))
	}

	// Check for error in response
	if eccResp.Error.Message != "" {
		return nil, fmt.Errorf("ServiceNow error: %s - %s", eccResp.Error.Message, eccResp.Error.Detail)
	}

	// Check if we got a valid SysID (indicates success)
	if eccResp.Result.SysID == "" {
		return nil, fmt.Errorf("ServiceNow error: No SysID returned in response")
	}

	return &eccResp, nil
}

func (c *Client) TestConnection() error {
	apiURL := fmt.Sprintf("%s://%s/api/now/table/sys_user", c.getProtocol(), c.instance)
	
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to ServiceNow: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ServiceNow connection test failed: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) GetInstanceURL() string {
	return fmt.Sprintf("%s://%s", c.getProtocol(), c.instance)
}

func (c *Client) getProtocol() string {
	if c.useHTTPS {
		return "https"
	}
	return "http"
}
