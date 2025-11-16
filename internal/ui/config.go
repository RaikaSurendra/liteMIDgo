package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ConfigModel struct {
	index     int
	focus     int
	questions []Question
	Answers   map[string]string
	quitting  bool
	width     int
	height    int
}

type Question struct {
	Key         string
	Text        string
	Default     string
	Password    bool
	Validator   func(string) error
	Placeholder string
}

type ConfigCompleteMsg map[string]string

func NewConfigModel() ConfigModel {
	questions := []Question{
		{
			Key:         "instance",
			Text:        "ServiceNow instance URL",
			Default:     "",
			Placeholder: "your-instance.service-now.com",
			Validator: func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("instance URL cannot be empty")
				}
				return nil
			},
		},
		{
			Key:         "username",
			Text:        "ServiceNow username",
			Default:     "",
			Placeholder: "your-username",
			Validator: func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("username cannot be empty")
				}
				return nil
			},
		},
		{
			Key:         "password",
			Text:        "ServiceNow password",
			Default:     "",
			Password:    true,
			Placeholder: "your-password",
			Validator: func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("password cannot be empty")
				}
				return nil
			},
		},
		{
			Key:         "host",
			Text:        "Server host",
			Default:     "0.0.0.0",
			Placeholder: "0.0.0.0",
			Validator: func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("host cannot be empty")
				}
				return nil
			},
		},
		{
			Key:         "port",
			Text:        "Server port",
			Default:     "8080",
			Placeholder: "8080",
			Validator: func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("port cannot be empty")
				}
				return nil
			},
		},
		{
			Key:         "use_https",
			Text:        "Use HTTPS?",
			Default:     "y",
			Placeholder: "y/n",
			Validator: func(s string) error {
				s = strings.ToLower(strings.TrimSpace(s))
				if s != "y" && s != "n" && s != "yes" && s != "no" {
					return fmt.Errorf("please enter y/n or yes/no")
				}
				return nil
			},
		},
		{
			Key:         "timeout",
			Text:        "Timeout (seconds)",
			Default:     "30",
			Placeholder: "30",
			Validator: func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("timeout cannot be empty")
				}
				return nil
			},
		},
	}

	answers := make(map[string]string)
	for _, q := range questions {
		answers[q.Key] = q.Default
	}

	return ConfigModel{
		questions: questions,
		Answers:   answers,
		index:     0,
		focus:     0,
	}
}

func (m ConfigModel) Init() tea.Cmd {
	return nil
}

func (m ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEnter, tea.KeyTab:
			currentQuestion := m.questions[m.index]
			answer := m.Answers[currentQuestion.Key]
			
			// Validate current answer
			if err := currentQuestion.Validator(answer); err != nil {
				return m, nil
			}

			if m.index >= len(m.questions)-1 {
				// Configuration complete
				m.quitting = true
				return m, tea.Quit
			}
			m.index++
			m.focus = m.index

		case tea.KeyShiftTab:
			if m.index > 0 {
				m.index--
				m.focus = m.index
			}

		case tea.KeyBackspace:
			if len(m.Answers[m.questions[m.index].Key]) > 0 {
				current := m.Answers[m.questions[m.index].Key]
				m.Answers[m.questions[m.index].Key] = current[:len(current)-1]
			}

		case tea.KeyRunes:
			m.Answers[m.questions[m.index].Key] += string(msg.Runes)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m ConfigModel) View() string {
	if m.quitting {
		return ""
	}

	// Styles
	var (
		titleStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7D56F4")).Padding(0, 2).Bold(true)
		focusedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Background(lipgloss.Color("#EEEDFF"))
		normalStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#A49BF5"))
		placeholderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B6B6B")).Italic(true)
		progressStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
		helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B6B6B")).Italic(true)
	)

	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("🔧 LiteMIDgo Configuration Setup"))
	content.WriteString("\n\n")

	// Progress indicator
	progress := fmt.Sprintf("[%d/%d]", m.index+1, len(m.questions))
	content.WriteString(progressStyle.Render(progress))
	content.WriteString("\n\n")

	// Questions
	for i, q := range m.questions {
		// Question text
		questionText := fmt.Sprintf("%s:", q.Text)
		if i == m.index {
			content.WriteString(focusedStyle.Render("▶ "+questionText))
		} else {
			content.WriteString(normalStyle.Render("  "+questionText))
		}
		content.WriteString("\n")

		// Answer field
		answer := m.Answers[q.Key]
		if q.Password && answer != "" {
			answer = strings.Repeat("•", len(answer))
		}
		
		if answer == "" && q.Placeholder != "" {
			content.WriteString(placeholderStyle.Render("  "+q.Placeholder))
		} else {
			if i == m.index {
				content.WriteString(focusedStyle.Render("  "+answer+"█"))
			} else {
				content.WriteString(normalStyle.Render("  "+answer))
			}
		}
		content.WriteString("\n\n")
	}

	// Help text
	content.WriteString(helpStyle.Render("↑/↓: Navigate • Enter: Next • Tab: Next • Shift+Tab: Previous • Ctrl+C: Quit"))

	return lipgloss.NewStyle().Padding(1, 2).Render(content.String())
}
