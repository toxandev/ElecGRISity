package pets

import (
	"telemetry-server/internal/config"
	"telemetry-server/internal/tui"

	"charm.land/huh/v2"
)

// RunForm displays a form to add or edit a pet configuration.
func RunForm(manager *config.ConfigManager, pet *config.PetConfig) bool {
	if pet.Type == "" {
		pet.Type = "pishock"
	}

	id := ""
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
						return "Shocker ID"
					}
					return "Lovense ID (ID)"
				}, &pet.Type).
				Value(&id),
			huh.NewInput().
				TitleFunc(func() string {
					if pet.Type == "pishock" {
						return "API Key"
					}
					return "Lovense IP (Secret)"
				}, &pet.Type).
				Value(&secret),
		),
	).WithTheme(tui.ResolveTheme(manager.Get().Theme)).
		WithKeyMap(tui.FormKeyMap())

	err := form.Run()
	tui.ClearScreen()
	if err == nil {
		if pet.Type == "pishock" {
			pet.LovenseID = ""
			pet.LovenseIP = ""
			pet.ShockerID = id
			pet.PiShockAPIKey = secret
		} else {
			pet.LovenseID = id
			pet.LovenseIP = secret
			pet.ShockerID = ""
			pet.PiShockAPIKey = ""
		}
		return true
	}
	return false
}
