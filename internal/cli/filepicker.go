package cli

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
)

type directoryPickerModel struct {
	filepicker filepicker.Model
	selected   string
	quitting   bool
}

func initialDirPickerModel(startDir string) directoryPickerModel {
	fp := filepicker.New()
	fp.DirAllowed = true
	fp.FileAllowed = false
	fp.ShowHidden = false

	if startDir != "" {
		if stat, err := os.Stat(startDir); err == nil && stat.IsDir() {
			fp.CurrentDirectory = startDir
		} else {
			dir, _ := os.Getwd()
			fp.CurrentDirectory = dir
		}
	} else {
		dir, _ := os.Getwd()
		fp.CurrentDirectory = dir
	}

	return directoryPickerModel{
		filepicker: fp,
	}
}

func (m directoryPickerModel) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m directoryPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	// In filepicker, didSelect, path := m.filepicker.DidSelectFile(msg) works for dirs if DirAllowed is true
	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		m.selected = path
		m.quitting = true
		return m, tea.Quit
	}

	return m, cmd
}

func (m directoryPickerModel) View() string {
	if m.quitting {
		return ""
	}

	var s string
	s += "\n  Select Game Directory (Press 'q' or 'esc' to cancel)\n\n"
	s += m.filepicker.View() + "\n"
	s += fmt.Sprintf("\n  Currently in: %s\n", m.filepicker.CurrentDirectory)

	return s
}

// RunDirPicker opens a terminal UI to pick a directory and returns the path.
// If the user cancels, it returns the original startDir.
func RunDirPicker(startDir string) string {
	m := initialDirPickerModel(startDir)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error running directory picker: %v\n", err)
		return startDir
	}

	if fm, ok := finalModel.(directoryPickerModel); ok {
		if fm.selected != "" {
			return fm.selected
		}
	}
	return startDir
}
