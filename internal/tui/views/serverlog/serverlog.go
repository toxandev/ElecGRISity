package serverlog

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"telemetry-server/internal/gamelink"
)

func formatLog(entry gamelink.LogEntry) string {
	var color string
	switch entry.Level {
	case gamelink.LogDebug:
		color = "#888888"
	case gamelink.LogInfo:
		color = "#04B575"
	case gamelink.LogWarn:
		color = "#FFAA00"
	case gamelink.LogError:
		color = "#FF4444"
	default:
		color = "#FFFFFF"
	}
	emoji := entry.ResolvedEmoji()
	msg := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(entry.Message)
	return fmt.Sprintf("%s %s", emoji, msg)
}

type logMsg gamelink.LogEntry

type Model struct {
	logs     []string
	logChan  <-chan gamelink.LogEntry
	minLevel gamelink.LogLevel
	quitting bool
	width    int
	height   int
	ready    bool
}

func NewModel(logChan <-chan gamelink.LogEntry, minLevel gamelink.LogLevel) Model {
	return Model{
		logs:     make([]string, 0),
		logChan:  logChan,
		minLevel: minLevel,
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
		entry, ok := <-m.logChan
		if !ok {
			return nil
		}
		return logMsg(entry)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	case logMsg:
		entry := gamelink.LogEntry(msg)
		if entry.Level >= m.minLevel {
			m.logs = append(m.logs, formatLog(entry))
		}
		return m, m.waitForLog()
	}

	return m, nil
}

func (m Model) View() tea.View {
	if m.quitting {
		v := tea.NewView("\nStopping server and returning to menu...\n")
		v.AltScreen = true
		return v
	}

	if !m.ready {
		v := tea.NewView("\nLoading...")
		v.AltScreen = true
		return v
	}

	titleStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1).
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4"))

	helpStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1).
		Foreground(lipgloss.Color("#04B575"))

	title := titleStyle.Render("⚡ May Gris lightning bless you all")
	help := helpStyle.Render("Server is running. Press 'q' or 'ctrl+c' to stop and return to the menu.")
	separator := strings.Repeat("─", m.width)

	banner := title + "\n" + help + "\n" + separator + "\n"
	bannerHeight := strings.Count(banner, "\n")

	var logsContent string
	if len(m.logs) == 0 {
		logsContent = "Waiting for game events..."
	} else {
		var sb strings.Builder
		for _, l := range m.logs {
			sb.WriteString("• ")
			sb.WriteString(l)
			sb.WriteString("\n")
		}
		logsContent = strings.TrimSuffix(sb.String(), "\n")
	}

	lines := strings.Split(logsContent, "\n")
	availableHeight := m.height - bannerHeight

	if len(lines) > availableHeight {
		lines = lines[len(lines)-availableHeight:]
	}

	for i, line := range lines {
		runes := []rune(line)
		if len(runes) > m.width {
			lines[i] = string(runes[:m.width])
		}
	}

	v := tea.NewView(banner + strings.Join(lines, "\n"))
	v.AltScreen = true
	return v
}
