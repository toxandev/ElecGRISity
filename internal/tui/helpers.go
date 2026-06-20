package tui

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
)

// ResolveTheme returns the huh Theme matching the config value.
func ResolveTheme(name string) huh.Theme {
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

// FormKeyMap returns a KeyMap with ESC and Ctrl+C bound to quit/abort.
func FormKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys("ctrl+c", "esc"))
	return km
}

// ClearScreen erases the terminal and moves the cursor to the top-left.
func ClearScreen() {
	fmt.Fprint(os.Stdout, "\033[H\033[2J\033[3J")
}
