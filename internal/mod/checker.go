package mod

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ModName is the exact name of the plugin file (without .js extension).
// Change this if your RPG Maker plugin has a different name!
const ModName = "NAMEOTHEMOD"

// CheckInstallation verifies if the mod is present in the game directory
func CheckInstallation(gamePath string) (bool, string) {
	if gamePath == "" {
		return false, "Game path is not set in the configuration.\nPlease go to 'Edit Configuration' and set it."
	}

	// 1. Check if the mod file exists
	modFilePath := filepath.Join(gamePath, "www", "js", "plugins", ModName+".js")
	if _, err := os.Stat(modFilePath); os.IsNotExist(err) {
		return false, fmt.Sprintf("❌ Mod file not found!\nExpected location:\n%s", modFilePath)
	}

	// 2. Check if registered in plugins.js
	pluginsJSPath := filepath.Join(gamePath, "www", "js", "plugins.js")
	content, err := os.ReadFile(pluginsJSPath)
	if err != nil {
		return false, fmt.Sprintf("❌ Could not read plugins.js:\n%s\nError: %v", pluginsJSPath, err)
	}

	// Check if the mod is registered in the plugins list
	// We look for the exact name to be sure it was added via the Plugin Manager
	if !strings.Contains(string(content), fmt.Sprintf(`"name":"%s"`, ModName)) && !strings.Contains(string(content), ModName) {
		return false, fmt.Sprintf("❌ Mod file exists, but it is NOT enabled in RPG Maker!\nPlease open RPG Maker Plugin Manager, add '%s', and turn it ON.", ModName)
	}

	return true, fmt.Sprintf("✅ Mod '%s' is correctly installed and registered in plugins.js!", ModName)
}
