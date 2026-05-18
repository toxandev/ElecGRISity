package main

import (
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"telemetry-server/config"
	"telemetry-server/internal/cli"
	"telemetry-server/internal/mod"
	"telemetry-server/internal/pet"
	"telemetry-server/internal/server"
)

const serverPort = 69420

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
						huh.NewOption("🔍 Check Mod Installation", "check"),
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
		} else if action == "check" {
			runModCheck(manager)
		} else if action == "start" {
			runServer(manager)
		}
	}
}

func runConfigMenu(manager *config.ConfigManager, cfgFile string) {
	cfg := manager.Get()

	logLevel := cfg.LogLevel
	pUser := cfg.PiShockUsername
	pKey := cfg.PiShockAPIKey
	pApp := cfg.PiShockAppName
	gPath := cfg.GamePath

	// Display configuration pane
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().Title("Log Level").Options(
				huh.NewOption("Debug", "debug"),
				huh.NewOption("Info", "info"),
				huh.NewOption("Warn", "warn"),
				huh.NewOption("Error", "error"),
			).Value(&logLevel),
			huh.NewInput().Title("RPG Maker Game Path (e.g. C:\\Games\\MyGame)").Value(&gPath),
			huh.NewInput().Title("PiShock Username").Value(&pUser),
			huh.NewInput().Title("PiShock API Key").Value(&pKey).EchoMode(huh.EchoModePassword),
			huh.NewInput().Title("PiShock App Name").Value(&pApp),
		),
	).WithTheme(huh.ThemeDracula())

	if err := form.Run(); err != nil {
		return // User cancelled or hit ESC
	}

	// Update the configuration safely
	err = manager.Update(func(c *config.Config) {
		c.LogLevel = logLevel
		c.PiShockUsername = pUser
		c.PiShockAPIKey = pKey
		c.PiShockAppName = pApp
		c.GamePath = gPath
	})

	if err != nil {
		fmt.Printf("\nValidation error: %v\n", err)
		return
	}

	// Save to config.yaml
	if err := manager.Save(cfgFile); err != nil {
		fmt.Printf("\nError saving config: %v\n", err)
	} else {
		fmt.Println("\nConfiguration saved successfully!")
	}
}

func runModCheck(manager *config.ConfigManager) {
	cfg := manager.Get()

	_, message := mod.CheckInstallation(cfg.GamePath)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Mod Installation Check").
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

	srv := server.NewServer(serverPort, pets, logChan)

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
	m := cli.NewModel(logChan)
	p := tea.NewProgram(m, tea.WithAltScreen())

	logChan <- fmt.Sprintf("Server initialized on port %d...", serverPort)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running UI: %v\n", err)
	}

	// Cancel context to close HTTP server once we exit Bubbletea loop
	cancel()
}
