package serverlog

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"telemetry-server/internal/tui/styles"
)

type logMsg string

type Model struct {
	logs     []string
	logChan  <-chan string
	quitting bool
}

func NewModel(logChan <-chan string) Model {
	return Model{
		logs:    make([]string, 0),
		logChan: logChan,
	}
}

func (m Model) Init() tea.Cmd {
	return m.waitForLog()
}

func (m Model) waitForLog() tea.Cmd {
	return func() tea.Msg {
		if m.logChan == nil {
			return nil
		}
		logText, ok := <-m.logChan
		if !ok {
			return nil
		}
		return logMsg(logText)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	case logMsg:
		m.logs = append(m.logs, string(msg))
		// Keep the view clean with max 10 logs
		if len(m.logs) > 10 {
			m.logs = m.logs[len(m.logs)-10:]
		}
		return m, m.waitForLog()
	}

	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return "\nStopping server and returning to menu...\n"
	}

	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(styles.Title.Render("⚡ Game Telemetry Server & PiShock"))
	b.WriteString("\n")
	b.WriteString(styles.Info.Render("Server is running. Press 'q' or 'ctrl+c' to stop and return to the menu."))
	b.WriteString("\n\n")

	var logsStr strings.Builder
	if len(m.logs) == 0 {
		logsStr.WriteString("Waiting for game events...\n")
	} else {
		for _, l := range m.logs {
			logsStr.WriteString(fmt.Sprintf("• %s\n", l))
		}
	}

	b.WriteString(styles.Log.Render(logsStr.String()))
	b.WriteString("\n")

	return b.String()
}
