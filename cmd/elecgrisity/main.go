package main

import (
	"context"
	"fmt"
	"log"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"telemetry-server/internal/config"
	"telemetry-server/internal/gamelink"
	"telemetry-server/internal/pet"
	"telemetry-server/internal/tui"

	configview "telemetry-server/internal/tui/views/config"
	installview "telemetry-server/internal/tui/views/install"
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
		theme := tui.ResolveTheme(manager.Get().Theme)
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
		).WithTheme(theme).WithKeyMap(tui.FormKeyMap())

		err := form.Run()
		tui.ClearScreen()
		if err != nil || action == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		if action == "config" {
			configview.RunMenu(manager, cfgFile)
		} else if action == "install" {
			installview.RunModCheck(manager)
		} else if action == "start" {
			runServer(manager)
		}
	}
}

func runServer(manager *config.ConfigManager) {
	cfg := manager.Get()

	// Channel for Bubbletea to display logs
	logChan := make(chan gamelink.LogEntry, 100)

	// Initialize Pets from Configuration
	pets := make(map[string]pet.Pet)
	for _, pc := range cfg.Pets {
		if pc.Type == "pishock" {
			pets[pc.Name] = &pet.PiShockPet{
				Name:      pc.Name,
				APIKey:    pc.PiShockAPIKey,
				ShockerID: pc.ShockerID,
			}
			maskedKey := "****"
			if len(pc.PiShockAPIKey) >= 4 {
				maskedKey = pc.PiShockAPIKey[0:4] + "****"
			}
   logChan <- gamelink.LogEntry{Level: gamelink.LogInfo, Emoji: "🐾", Message: fmt.Sprintf("Initialized PiShock pet: %s, ShockerID: %s, APIKey: %s", pc.Name, pc.ShockerID, maskedKey)}
		} else if pc.Type == "lovense" {
			pets[pc.Name] = &pet.LovensePet{
				Name:      pc.Name,
				LovenseID: pc.LovenseID,
				LovenseIP: pc.LovenseIP,
			}
		}
	}

	srv := gamelink.NewServer(serverPort, pets, logChan)

	// Context for graceful shutdown when we exit Bubbletea
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start web server asynchronously
	go func() {
		if err := srv.Start(ctx); err != nil {
			logChan <- gamelink.LogEntry{Level: gamelink.LogError, Message: fmt.Sprintf("Server stopped: %v", err)}
		}
	}()

	// Start Bubbletea UI
	minLevel := gamelink.ParseLogLevel(cfg.LogLevel)
	m := serverlog.NewModel(logChan, minLevel)
	p := tea.NewProgram(m)

 logChan <- gamelink.LogEntry{Level: gamelink.LogInfo, Emoji: "⚡", Message: fmt.Sprintf("Server initialized on port %d...", serverPort)}

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running UI: %v\n", err)
	}

	// Cancel context to close HTTP server once we exit Bubbletea loop
	cancel()
}
