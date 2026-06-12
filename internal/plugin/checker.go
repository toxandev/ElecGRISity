package plugin

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
const ModName = "elecgrisity"

var ErrAlreadyInstalled = fmt.Errorf("mod is already installed")

// InstallMod verifies if the binary is next to the game folder, and automatically installs and registers the plugin.
func InstallMod() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)

	gameFolder := "Click Me"

	// 1. Check if we are next to Click Me/BLOODMONEY!.exe
	gameExePath := filepath.Join(exeDir, gameFolder, "BLOODMONEY!.exe")
	if _, err := os.Stat(gameExePath); os.IsNotExist(err) {
		return fmt.Errorf("game not found. Expected to find it at: %s\nPlease place this program next to the '%s' folder", gameExePath, gameFolder)
	}

	// 2. Install/Update the mod file
	modFilePath := filepath.Join(exeDir, gameFolder, "www", "js", "plugins", ModName+".js")
	if err := os.MkdirAll(filepath.Dir(modFilePath), 0755); err != nil {
		return fmt.Errorf("could not create mod directory: %v", err)
	}
	if err := os.WriteFile(modFilePath, modScript, 0644); err != nil {
		return fmt.Errorf("could not write mod file: %v", err)
	}

	// 3. Check if registered in plugins.js
	pluginsJSPath := filepath.Join(exeDir, gameFolder, "www", "js", "plugins.js")
	content, err := os.ReadFile(pluginsJSPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create a basic plugins.js
			content = []byte(fmt.Sprintf("var $plugins = [\n{\"name\":\"%s\",\"status\":true,\"description\":\"Smart Event Sender\",\"parameters\":{}}\n];", ModName))
			if err := os.WriteFile(pluginsJSPath, content, 0644); err != nil {
				return fmt.Errorf("could not create plugins.js: %v", err)
			}
		} else {
			return fmt.Errorf("could not read plugins.js: %v", err)
		}
	} else {
		// Check if the mod is registered in the plugins list
		if !strings.Contains(string(content), fmt.Sprintf(`"name":"%s"`, ModName)) && !strings.Contains(string(content), ModName) {
			// Auto-register it
			newPlugin := fmt.Sprintf(",\n{\"name\":\"%s\",\"status\":true,\"description\":\"Smart Event Sender\",\"parameters\":{}}\n];", ModName)
			newContent := strings.Replace(string(content), "\n];", newPlugin, 1)
			if newContent == string(content) {
				newContent = strings.Replace(string(content), "];", newPlugin[1:], 1)
			}
			if err := os.WriteFile(pluginsJSPath, []byte(newContent), 0644); err != nil {
				return fmt.Errorf("could not update plugins.js: %v", err)
			}
		} else {
			return ErrAlreadyInstalled
		}
	}

	return nil
}
