package pets

import (
	"fmt"
	"strings"

	"telemetry-server/internal/config"
	"telemetry-server/internal/tui/styles"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Model struct {
	Manager     *config.ConfigManager
	Pets        []config.PetConfig
	cursor      int    // 0 to len(pets) (last is add)
	subCursor   int    // 0 for Edit, 1 for Delete
	action      string // "edit", "delete", "add", "back", ""
	selectedIdx int
}

func NewModel(manager *config.ConfigManager) Model {
	return Model{
		Manager: manager,
		Pets:    manager.Get().Pets,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.action = "back"
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.subCursor = 0
			}
		case "down", "j":
			if m.cursor < len(m.Pets) {
				m.cursor++
				m.subCursor = 0
			}
		case "left", "h":
			if m.cursor < len(m.Pets) && m.subCursor > 0 {
				m.subCursor--
			}
		case "right", "l":
			if m.cursor < len(m.Pets) && m.subCursor < 1 {
				m.subCursor++
			}
		case "enter", " ":
			if m.cursor == len(m.Pets) {
				m.action = "add"
			} else {
				m.selectedIdx = m.cursor
				if m.subCursor == 0 {
					m.action = "edit"
				} else {
					m.action = "delete"
				}
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() tea.View {
	var b strings.Builder

	b.WriteString("\n" + styles.Title.Render("🐾 Manage Pets") + "\n\n")

	for i, pet := range m.Pets {
		var content strings.Builder
		content.WriteString(lipgloss.NewStyle().Bold(true).Render(pet.Name) + "\n")
		content.WriteString(fmt.Sprintf("Type: %s\n", pet.Type))

		if pet.Type == "pishock" {
			content.WriteString(fmt.Sprintf("PiShock API Key: %s\nID: %s\n", pet.PiShockAPIKey[0:4]+"****", pet.ShockerID))
		} else if pet.Type == "lovense" {
			content.WriteString(fmt.Sprintf("Lovense ID: %s\nIP: %s\n", pet.LovenseID, pet.LovenseIP))
		}

		content.WriteString("\n")

		editBtn := styles.Button.Render("Edit")
		deleteBtn := styles.DeleteButton.Render("Delete")

		if m.cursor == i {
			if m.subCursor == 0 {
				editBtn = styles.SelectedButton.Render("Edit")
			} else {
				deleteBtn = styles.SelectedDeleteButton.Render("Delete")
			}
			content.WriteString(editBtn + deleteBtn)
			b.WriteString(styles.SelectedCard.Render(content.String()) + "\n")
		} else {
			content.WriteString(editBtn + deleteBtn)
			b.WriteString(styles.Card.Render(content.String()) + "\n")
		}
	}

	addText := "➕ Add New Pet"
	if m.cursor == len(m.Pets) {
		b.WriteString(styles.SelectedAddCard.Render(addText) + "\n")
	} else {
		b.WriteString(styles.AddCard.Render(addText) + "\n")
	}

	b.WriteString("\n" + styles.Info.Render("↑/↓: Select Pet • ←/→: Select Action • Enter: Confirm • Esc/q: Back"))
	b.WriteString("\n")

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

func (m Model) GetAction() string {
	return m.action
}

func (m Model) GetSelectedIdx() int {
	return m.selectedIdx
}
