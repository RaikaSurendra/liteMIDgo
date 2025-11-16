package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"litemidgo/config"
	"litemidgo/internal/server"
)

type ServerDashboardModel struct {
	server      *server.Server
	config      *config.Config
	status      ServerStatus
	startTime   time.Time
	requests    int
	lastUpdate  time.Time
	spinner     int
	width       int
	height      int
	quitting    bool
}

type ServerStatus int

const (
	StatusStopped ServerStatus = iota
	StatusStarting
	StatusRunning
	StatusError
)

func NewServerDashboardModel(cfg *config.Config) ServerDashboardModel {
	return ServerDashboardModel{
		config:     cfg,
		status:     StatusStopped,
		startTime:  time.Now(),
		requests:   0,
		lastUpdate: time.Now(),
		spinner:    0,
	}
}

type ServerStatusMsg struct {
	status  ServerStatus
	error   error
	requests int
}

func (m ServerDashboardModel) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}


func (m ServerDashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			if m.server != nil {
				// Stop the server gracefully
				m.server.Stop()
			}
			return m, tea.Quit

		case tea.KeyEnter, tea.KeySpace:
			if m.status == StatusStopped {
				m.status = StatusStarting
				m.server = server.NewServer(m.config)
				return m, tea.Batch(
					startServer(m.server),
					tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
						return TickMsg(t)
					}),
				)
			} else if m.status == StatusRunning {
				m.status = StatusStopped
				if m.server != nil {
					m.server.Stop()
					m.server = nil
				}
			}
		case tea.KeyRunes:
			if strings.ToLower(string(msg.Runes)) == "q" {
				m.quitting = true
				if m.server != nil {
					m.server.Stop()
				}
				return m, tea.Quit
			}
		}

	case TickMsg:
		m.spinner = (m.spinner + 1) % 4
		m.lastUpdate = time.Now()
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return TickMsg(t)
		})

	case ServerStatusMsg:
		m.status = msg.status
		if msg.error != nil {
			// Handle server error
		}
		if msg.requests > m.requests {
			m.requests = msg.requests
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func startServer(srv *server.Server) tea.Cmd {
	return func() tea.Msg {
		go func() {
			if err := srv.Start(); err != nil {
				// Send error message
			}
		}()
		return ServerStatusMsg{status: StatusRunning}
	}
}

func (m ServerDashboardModel) View() string {
	// Styles
	var (
		titleStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7D56F4")).Padding(0, 2).Bold(true)
		headerStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#5A47E8")).Padding(0, 1)
		successStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ADE80")).Bold(true)
		errorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Bold(true)
		warningStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24")).Bold(true)
		normalStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#A49BF5"))
		infoStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA"))
		helpStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B6B6B")).Italic(true)
		boxStyle         = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#5A47E8")).Padding(1)
		spinnerChars     = []string{"⠋", "⠙", "⠹", "⠸"}
	)

	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("🚀 LiteMIDgo Server Dashboard"))
	content.WriteString("\n\n")

	// Server Status Box
	statusText := "Stopped"
	statusColor := errorStyle
	switch m.status {
	case StatusStarting:
		spinner := spinnerChars[m.spinner]
		statusText = fmt.Sprintf("%s Starting", spinner)
		statusColor = warningStyle
	case StatusRunning:
		statusText = "Running"
		statusColor = successStyle
	case StatusError:
		statusText = "Error"
		statusColor = errorStyle
	}

	statusBox := fmt.Sprintf("Status: %s", statusColor.Render(statusText))
	content.WriteString(boxStyle.Render(headerStyle.Render("Server Status") + "\n" + statusBox))
	content.WriteString("\n\n")

	// Configuration Box
	configBox := fmt.Sprintf(
		"Host: %s\nPort: %d\nInstance: %s",
		normalStyle.Render(m.config.Server.Host),
		normalStyle.Render(fmt.Sprintf("%d", m.config.Server.Port)),
		normalStyle.Render(m.config.ServiceNow.Instance),
	)
	content.WriteString(boxStyle.Render(headerStyle.Render("Configuration") + "\n" + configBox))
	content.WriteString("\n\n")

	// Statistics Box
	uptime := time.Since(m.startTime)
	statsBox := fmt.Sprintf(
		"Uptime: %s\nRequests: %d\nLast Update: %s",
		normalStyle.Render(uptime.Round(time.Second).String()),
		normalStyle.Render(fmt.Sprintf("%d", m.requests)),
		normalStyle.Render(m.lastUpdate.Format("15:04:05")),
	)
	content.WriteString(boxStyle.Render(headerStyle.Render("Statistics") + "\n" + statsBox))
	content.WriteString("\n\n")

	// Endpoints Box
	endpointsBox := fmt.Sprintf(
		"GET  /health\nPOST /proxy/ecc_queue\nGET  /",
		infoStyle.Render("GET  /health\nPOST /proxy/ecc_queue\nGET  /"),
	)
	content.WriteString(boxStyle.Render(headerStyle.Render("Available Endpoints") + "\n" + endpointsBox))

	// Help text
	content.WriteString("\n\n")
	if m.status == StatusStopped {
		content.WriteString(helpStyle.Render("Press Enter/Space to start server • Q/Ctrl+C to exit"))
	} else if m.status == StatusRunning {
		content.WriteString(helpStyle.Render("Press Enter/Space to stop server • Q/Ctrl+C to exit"))
	} else {
		content.WriteString(helpStyle.Render("Q/Ctrl+C to exit"))
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(content.String())
}
