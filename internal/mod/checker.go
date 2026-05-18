package mod

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed WalletLogger.js
var modScript []byte

// ModName is the exact name of the plugin file (without .js extension).
const ModName = "WalletLogger"

// CheckInstallation verifies if the binary is next to the clickme folder and installs the plugin
func CheckInstallation() (bool, string) {
	exePath, err := os.Executable()
	if err != nil {
		return false, fmt.Sprintf("❌ Could not determine executable path:\n%v", err)
	}
	exeDir := filepath.Dir(exePath)

	// 1. Check if we are next to clickme/BLOODMONEY.exe
	gameExePath := filepath.Join(exeDir, "clickme", "BLOODMONEY.exe")
	if _, err := os.Stat(gameExePath); os.IsNotExist(err) {
		return false, fmt.Sprintf("❌ Game not found!\nExpected to find:\n`%s`\nPlease place this program next to the 'clickme' folder.", gameExePath)
	}

	// 2. Install/Update the mod file
	modFilePath := filepath.Join(exeDir, "clickme", "www", "js", "plugins", ModName+".js")
	if err := os.WriteFile(modFilePath, modScript, 0644); err != nil {
		return false, fmt.Sprintf("❌ Could not install mod file:\n%v", err)
	}

	// 3. Check if registered in plugins.js
	pluginsJSPath := filepath.Join(exeDir, "clickme", "www", "js", "plugins.js")
	content, err := os.ReadFile(pluginsJSPath)
	if err != nil {
		return false, fmt.Sprintf("❌ Could not read plugins.js:\n`%s`\nError: %v", pluginsJSPath, err)
	}

	// Check if the mod is registered in the plugins list
	if !strings.Contains(string(content), fmt.Sprintf(`"name":"%s"`, ModName)) && !strings.Contains(string(content), ModName) {
		return false, fmt.Sprintf("❌ Mod file is installed, but it is NOT enabled in RPG Maker!\nPlease open RPG Maker Plugin Manager, add '%s', and turn it ON. Or add it to plugins.js manually.", ModName)
	}

	return true, fmt.Sprintf("✅ Mod '%s' is correctly installed and registered in plugins.js!", ModName)
}
