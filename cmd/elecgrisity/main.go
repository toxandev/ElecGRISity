package main

import (
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"telemetry-server/internal/config"
	"telemetry-server/internal/pet"
	"telemetry-server/internal/plugin"
	"telemetry-server/internal/telemetry"

	petsview "telemetry-server/internal/tui/views/pets"
	serverlog "telemetry-server/internal/tui/views/serverlog"
)

const serverPort = 42069

func main() {
	cfgFile := "config.yaml"
	manager := config.NewConfigManager()

	if err := manager.Load(cfgFile); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	for {
		var action string

		// Main Menu
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("⚡ PiShock Game Telemetry").
					Description("Choose an action:").
					Options(
						huh.NewOption("⏯️ Start Server", "start"),
						huh.NewOption("🔍 Install Mod", "install"),
						huh.NewOption("⚙ Edit Configuration", "config"),
						huh.NewOption("✖ Exit", "exit"),
					).
					Value(&action),
			),
		)

		err := form.Run()
		if err != nil || action == "exit" {
			fmt.Println("\nGoodbye!")
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
		)

		if err := form.Run(); err != nil || action == "cancel" {
			return
		}

		if action == "save" {
			if err := manager.Save(cfgFile); err != nil {
				fmt.Printf("\nError saving config: %v\n", err)
			} else {
				fmt.Println("\nConfiguration saved successfully!")
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
			huh.NewInput().Title("PiShock Username").Value(&pUser),
			huh.NewInput().Title("PiShock API Key").Value(&pKey).EchoMode(huh.EchoModePassword),
			huh.NewInput().Title("PiShock App Name").Value(&pApp),
		),
	).WithTheme(huh.ThemeDracula())

	if err := form.Run(); err == nil {
		manager.Update(func(c *config.Config) {
			c.LogLevel = logLevel
			c.PiShockUsername = pUser
			c.PiShockAPIKey = pKey
			c.PiShockAppName = pApp
		})
	}
}

func editPetsConfig(manager *config.ConfigManager) {
	for {
		m := petsview.NewModel(manager)
		p := tea.NewProgram(m, tea.WithAltScreen())

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
			if runPetForm(&newPet) {
				manager.Update(func(c *config.Config) {
					c.Pets = append(c.Pets, newPet)
				})
			}
		} else if action == "edit" {
			idx := petsModel.GetSelectedIdx()
			cfg := manager.Get()
			if idx >= 0 && idx < len(cfg.Pets) {
				petToEdit := cfg.Pets[idx]
				if runPetForm(&petToEdit) {
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

func runPetForm(pet *config.PetConfig) bool {
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
	).WithTheme(huh.ThemeDracula())

	err := form.Run()
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
	if err != nil {
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
	).WithTheme(huh.ThemeDracula())

	form.Run()
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
	p := tea.NewProgram(m, tea.WithAltScreen())

	logChan <- fmt.Sprintf("Server initialized on port %d...", serverPort)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running UI: %v\n", err)
	}

	// Cancel context to close HTTP server once we exit Bubbletea loop
	cancel()
}
