package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"litemidgo/config"
	"litemidgo/internal/servicenow"
)

type ConnectionTestModel struct {
	status    TestStatus
	instance  string
	error     error
	spinner   int
	width     int
	height    int
	quitting  bool
}

type TestStatus int

const (
	StatusIdle TestStatus = iota
	StatusTesting
	StatusSuccess
	StatusFailed
)

func NewConnectionTestModel(instance string) ConnectionTestModel {
	return ConnectionTestModel{
		status:   StatusIdle,
		instance: instance,
		spinner:  0,
	}
}

type TestCompleteMsg struct {
	success bool
	error   error
}

func (m ConnectionTestModel) Init() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

type TickMsg time.Time

func (m ConnectionTestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc, tea.KeyEnter, tea.KeySpace:
			m.quitting = true
			return m, tea.Quit
		}

	case TickMsg:
		if m.status == StatusTesting {
			m.spinner = (m.spinner + 1) % 4
			return m, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
				return TickMsg(t)
			})
		} else if m.status == StatusIdle {
			m.status = StatusTesting
			return m, tea.Batch(
				tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
					return TickMsg(t)
				}),
				performConnectionTest(m.instance),
			)
		}

	case TestCompleteMsg:
		if msg.success {
			m.status = StatusSuccess
		} else {
			m.status = StatusFailed
			m.error = msg.error
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func performConnectionTest(instance string) tea.Cmd {
	return func() tea.Msg {
		// Load configuration
		cfg, err := config.LoadConfig("")
		if err != nil {
			return TestCompleteMsg{success: false, error: err}
		}

		// Create ServiceNow client and test connection
		snowClient := servicenow.NewClient(&cfg.ServiceNow)
		err = snowClient.TestConnection()
		
		return TestCompleteMsg{
			success: err == nil,
			error:   err,
		}
	}
}

func (m ConnectionTestModel) View() string {
	// Styles
	var (
		titleStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7D56F4")).Padding(0, 2).Bold(true)
		successStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ADE80")).Bold(true)
		errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Bold(true)
		testingStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA"))
		normalStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#A49BF5"))
		helpStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B6B6B")).Italic(true)
		spinnerChars   = []string{"⠋", "⠙", "⠹", "⠸"}
	)

	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("🔍 ServiceNow Connection Test"))
	content.WriteString("\n\n")

	// Status based on current state
	switch m.status {
	case StatusIdle:
		content.WriteString(normalStyle.Render("Preparing to test connection..."))

	case StatusTesting:
		spinner := spinnerChars[m.spinner]
		content.WriteString(testingStyle.Render(fmt.Sprintf("%s Testing connection to %s...", spinner, m.instance)))
		content.WriteString("\n\n")
		content.WriteString(normalStyle.Render("This may take a few moments..."))

	case StatusSuccess:
		content.WriteString(successStyle.Render("✅ Connection Successful!"))
		content.WriteString("\n\n")
		content.WriteString(normalStyle.Render(fmt.Sprintf("Successfully connected to %s", m.instance)))
		content.WriteString("\n\n")
		content.WriteString(successStyle.Render("Your ServiceNow instance is accessible and credentials are valid."))

	case StatusFailed:
		content.WriteString(errorStyle.Render("❌ Connection Failed"))
		content.WriteString("\n\n")
		if m.error != nil {
			content.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.error)))
			content.WriteString("\n\n")
		}
		content.WriteString(normalStyle.Render("Please check:"))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render("• Instance URL is correct"))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render("• Username and password are valid"))
		content.WriteString("\n")
		content.WriteString(normalStyle.Render("• Network connectivity to ServiceNow"))
	}

	// Help text
	if m.status == StatusSuccess || m.status == StatusFailed {
		content.WriteString("\n\n")
		content.WriteString(helpStyle.Render("Press Enter, Space, or Escape to exit"))
	}

	return lipgloss.NewStyle().Padding(2, 3).Render(content.String())
}
