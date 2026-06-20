package installview

import (
	"errors"
	"fmt"

	"telemetry-server/internal/config"
	"telemetry-server/internal/plugin"
	"telemetry-server/internal/tui"

	"charm.land/huh/v2"
)

// RunModCheck checks if the mod is installed and offers to install it if not.
func RunModCheck(manager *config.ConfigManager) {
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
	).WithTheme(tui.ResolveTheme(manager.Get().Theme)).
		WithKeyMap(tui.FormKeyMap())

	form.Run()
	tui.ClearScreen()
}
