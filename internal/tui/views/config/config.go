package configview

import (
	"fmt"

	"telemetry-server/internal/config"
	"telemetry-server/internal/tui"
	petsview "telemetry-server/internal/tui/views/pets"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

// RunMenu displays the main configuration menu.
func RunMenu(manager *config.ConfigManager, cfgFile string) {
	for {
		var action string
		theme := tui.ResolveTheme(manager.Get().Theme)
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Configuration Menu").
					Description("Choose an option to configure:").
					Options(
						huh.NewOption("General Settings", "general"),
						huh.NewOption("Manage Pets", "pets"),
						huh.NewOption("Save & Return", "save"),
						huh.NewOption("Cancel", "cancel"),
					).
					Value(&action),
			),
		).WithTheme(theme).WithKeyMap(tui.FormKeyMap())

		err := form.Run()
		tui.ClearScreen()
		if err != nil || action == "cancel" {
			return
		}

		if action == "save" {
			if err := manager.Save(cfgFile); err != nil {
				fmt.Printf("Error saving config: %v\n", err)
			} else {
				fmt.Println("Configuration saved successfully!")
			}
			return
		}

		if action == "general" {
			editGeneralConfig(manager)
		} else if action == "pets" {
			editPetsConfig(manager)
		}
	}
}

func editGeneralConfig(manager *config.ConfigManager) {
	cfg := manager.Get()
	logLevel := cfg.LogLevel
	themeName := cfg.Theme

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().Title("Log Level").Options(
				huh.NewOption("Debug", "debug"),
				huh.NewOption("Info", "info"),
				huh.NewOption("Warn", "warn"),
				huh.NewOption("Error", "error"),
			).Value(&logLevel),
			huh.NewSelect[string]().Title("Theme").Options(
				huh.NewOption("Base", "base"),
				huh.NewOption("Base16", "base16"),
				huh.NewOption("Catppuccin", "catppuccin"),
				huh.NewOption("Charm", "charm"),
				huh.NewOption("Dracula", "dracula"),
			).Value(&themeName),
		),
	).WithTheme(tui.ResolveTheme(cfg.Theme)).
		WithKeyMap(tui.FormKeyMap())

	err := form.Run()
	tui.ClearScreen()
	if err == nil {
		manager.Update(func(c *config.Config) {
			c.LogLevel = logLevel
			c.Theme = themeName
		})
	}
}

func editPetsConfig(manager *config.ConfigManager) {
	for {
		m := petsview.NewModel(manager)
		p := tea.NewProgram(m)

		mod, err := p.Run()
		if err != nil {
			fmt.Printf("Error running pet menu: %v\n", err)
			return
		}

		petsModel, ok := mod.(petsview.Model)
		if !ok || petsModel.GetAction() == "back" || petsModel.GetAction() == "" {
			return
		}

		action := petsModel.GetAction()

		if action == "add" {
			newPet := config.PetConfig{Type: "pishock"}
			if petsview.RunForm(manager, &newPet) {
				manager.Update(func(c *config.Config) {
					c.Pets = append(c.Pets, newPet)
				})
			}
		} else if action == "edit" {
			idx := petsModel.GetSelectedIdx()
			cfg := manager.Get()
			if idx >= 0 && idx < len(cfg.Pets) {
				petToEdit := cfg.Pets[idx]
				if petsview.RunForm(manager, &petToEdit) {
					manager.Update(func(c *config.Config) {
						c.Pets[idx] = petToEdit
					})
				}
			}
		} else if action == "delete" {
			idx := petsModel.GetSelectedIdx()
			cfg := manager.Get()
			if idx >= 0 && idx < len(cfg.Pets) {
				manager.Update(func(c *config.Config) {
					c.Pets = append(c.Pets[:idx], c.Pets[idx+1:]...)
				})
			}
		}
	}
}
