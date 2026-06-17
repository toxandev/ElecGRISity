package serverlog

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type logMsg string

type Model struct {
	logs     []string
	logChan  <-chan string
	quitting bool
	width    int
	height   int
	ready    bool
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
		m.logs = append(m.logs, string(msg))
		return m, m.waitForLog()
	}

	return m, nil
}

func (m Model) View() tea.View {
	if m.quitting {
		return tea.NewView("\nStopping server and returning to menu...\n")
	}

	if !m.ready {
		return tea.NewView("\nLoading...")
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

	title := titleStyle.Render("⚡ May Gris thunder bless you all")
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

	return tea.NewView(banner + strings.Join(lines, "\n"))
}
