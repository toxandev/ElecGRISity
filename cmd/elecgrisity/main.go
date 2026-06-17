package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"telemetry-server/internal/config"
	"telemetry-server/internal/pet"
	"telemetry-server/internal/plugin"
	"telemetry-server/internal/telemetry"

	petsview "telemetry-server/internal/tui/views/pets"
	serverlog "telemetry-server/internal/tui/views/serverlog"
)

const serverPort = 42069

// formKeyMap returns a KeyMap with ESC and Ctrl+C bound to quit/abort.
func formKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys("ctrl+c", "esc"))
	return km
}

// clearScreen erases the terminal and moves the cursor to the top-left.
// Called after every huh form to prevent inline rendering artifacts.
func clearScreen() {
	fmt.Fprint(os.Stdout, "\033[H\033[2J\033[3J")
}

// resolveTheme returns the huh Theme matching the config value.
func resolveTheme(name string) huh.Theme {
	switch name {
	case "base16":
		return huh.ThemeFunc(huh.ThemeBase16)
	case "catppuccin":
		return huh.ThemeFunc(huh.ThemeCatppuccin)
	case "charm":
		return huh.ThemeFunc(huh.ThemeCharm)
	case "dracula":
		return huh.ThemeFunc(huh.ThemeDracula)
	default:
		return huh.ThemeFunc(huh.ThemeBase)
	}
}

func main() {
	cfgFile := "config.yaml"
	manager := config.NewConfigManager()

	if err := manager.Load(cfgFile); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	for {
		var action string

		// Main Menu
		theme := resolveTheme(manager.Get().Theme)
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("⚡Elecgrisity⚡").
					Description("Choose an action:").
					Options(
						huh.NewOption("⏯️ Start Server", "start"),
						huh.NewOption("🔍 Install Mod", "install"),
						huh.NewOption("⚙ Edit Configuration", "config"),
						huh.NewOption("✖ Exit", "exit"),
					).
					Value(&action),
			),
		).WithTheme(theme).WithKeyMap(formKeyMap())

		err := form.Run()
		clearScreen()
		if err != nil || action == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		if action == "config" {
			runConfigMenu(manager, cfgFile)
		} else if action == "install" {
			runModCheck(manager)
		} else if action == "start" {
			runServer(manager)
		}
	}
}

func runConfigMenu(manager *config.ConfigManager, cfgFile string) {
	for {
		var action string
		theme := resolveTheme(manager.Get().Theme)
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
		).WithTheme(theme).WithKeyMap(formKeyMap())

		err := form.Run()
		clearScreen()
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
	pUser := cfg.PiShockUsername
	pKey := cfg.PiShockAPIKey
	pApp := cfg.PiShockAppName

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
			huh.NewInput().Title("PiShock Username").Value(&pUser),
			huh.NewInput().Title("PiShock API Key").Value(&pKey).EchoMode(huh.EchoModePassword),
			huh.NewInput().Title("PiShock App Name").Value(&pApp),
		),
	).WithTheme(resolveTheme(cfg.Theme)).
		WithKeyMap(formKeyMap())

	err := form.Run()
	clearScreen()
	if err == nil {
		manager.Update(func(c *config.Config) {
			c.LogLevel = logLevel
			c.Theme = themeName
			c.PiShockUsername = pUser
			c.PiShockAPIKey = pKey
			c.PiShockAppName = pApp
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
			if runPetForm(manager, &newPet) {
				manager.Update(func(c *config.Config) {
					c.Pets = append(c.Pets, newPet)
				})
			}
		} else if action == "edit" {
			idx := petsModel.GetSelectedIdx()
			cfg := manager.Get()
			if idx >= 0 && idx < len(cfg.Pets) {
				petToEdit := cfg.Pets[idx]
				if runPetForm(manager, &petToEdit) {
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

func runPetForm(manager *config.ConfigManager, pet *config.PetConfig) bool {
	if pet.Type == "" {
		pet.Type = "pishock"
	}

	id := pet.ShareCode
	if pet.Type == "lovense" {
		id = pet.LovenseID
	}
	secret := pet.LovenseIP // Use LovenseIP for the 'secret' field

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Name").Value(&pet.Name),
			huh.NewSelect[string]().
				Title("Type").
				Options(
					huh.NewOption("PiShock", "pishock"),
					huh.NewOption("Lovense", "lovense"),
				).
				Value(&pet.Type),
		),
		huh.NewGroup(
			huh.NewInput().
				TitleFunc(func() string {
					if pet.Type == "pishock" {
						return "Share Code (ID)"
					}
					return "Lovense ID (ID)"
				}, &pet.Type).
				Value(&id),
			huh.NewInput().
				TitleFunc(func() string {
					if pet.Type == "pishock" {
						return "Secret (Not used for PiShock)"
					}
					return "Lovense IP (Secret)"
				}, &pet.Type).
				Value(&secret),
		),
	).WithTheme(resolveTheme(manager.Get().Theme)).
		WithKeyMap(formKeyMap())

	err := form.Run()
	clearScreen()
	if err == nil {
		if pet.Type == "pishock" {
			pet.ShareCode = id
			pet.LovenseID = ""
			pet.LovenseIP = ""
		} else {
			pet.LovenseID = id
			pet.LovenseIP = secret
			pet.ShareCode = ""
		}
		return true
	}
	return false
}

func runModCheck(manager *config.ConfigManager) {
	err := plugin.InstallMod()

	var message string
	if errors.Is(err, plugin.ErrAlreadyInstalled) {
		message = "✅ Mod is already installed"
	} else if err != nil {
		message = fmt.Sprintf("❌ Error installing mod:\n%v", err)
	} else {
		message = "✅ Mod successfully installed!"
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Mod Installation").
				Description(message),
		),
	).WithTheme(resolveTheme(manager.Get().Theme)).
		WithKeyMap(formKeyMap())

	form.Run()
	clearScreen()
}

func runServer(manager *config.ConfigManager) {
	cfg := manager.Get()

	// Channel for Bubbletea to display logs
	logChan := make(chan string, 100)

	// Initialize Pets from Configuration
	pets := make(map[string]pet.Pet)
	for _, pc := range cfg.Pets {
		if pc.Type == "pishock" {
			pets[pc.Name] = &pet.PiShockPet{
				Name:      pc.Name,
				ShareCode: pc.ShareCode,
				Username:  cfg.PiShockUsername,
				APIKey:    cfg.PiShockAPIKey,
				AppName:   cfg.PiShockAppName,
			}
		} else if pc.Type == "lovense" {
			pets[pc.Name] = &pet.LovensePet{
				Name:      pc.Name,
				LovenseID: pc.LovenseID,
				LovenseIP: pc.LovenseIP,
			}
		}
	}

	srv := telemetry.NewServer(serverPort, pets, logChan)

	// Context for graceful shutdown when we exit Bubbletea
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start web server asynchronously
	go func() {
		if err := srv.Start(ctx); err != nil {
			logChan <- fmt.Sprintf("Server stopped: %v", err)
		}
	}()

	// Start Bubbletea UI
	m := serverlog.NewModel(logChan)
	p := tea.NewProgram(m)

	logChan <- fmt.Sprintf("Server initialized on port %d...", serverPort)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running UI: %v\n", err)
	}

	// Cancel context to close HTTP server once we exit Bubbletea loop
	cancel()
}
