package styles

import "charm.land/lipgloss/v2"

var (
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		MarginBottom(1)

	Info = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		MarginBottom(1)

	Log = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#383838")).
		Padding(1, 2).
		Width(70)

	Card = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2).
		MarginBottom(1).
		Width(40)

	SelectedCard = Card.
			BorderForeground(lipgloss.Color("#FF76B8")).
			Background(lipgloss.Color("#2A2A2A"))

	AddCard = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(lipgloss.Color("#04B575")).
		Padding(1, 2).
		MarginBottom(1).
		Width(40).
		Align(lipgloss.Center)

	SelectedAddCard = AddCard.
			BorderForeground(lipgloss.Color("#04B575")).
			Background(lipgloss.Color("#2A2A2A")).
			Bold(true)

	Button = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFF")).
		Background(lipgloss.Color("#444")).
		Padding(0, 1).
		MarginRight(1)

	SelectedButton = Button.
			Background(lipgloss.Color("#FF76B8")).
			Bold(true)

	DeleteButton = Button.
			Background(lipgloss.Color("#E03131"))

	SelectedDeleteButton = DeleteButton.
				Bold(true).
				Foreground(lipgloss.Color("#FFF")).
				Background(lipgloss.Color("#FF4B4B"))
)
