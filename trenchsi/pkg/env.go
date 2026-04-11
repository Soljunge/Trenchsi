// all environment variables including default values put here

package pkg

import (
	"os"
	"path/filepath"
)

const (
	Logo = "🪖"
	// AppName is the name of the app
	AppName = "Trenchsi"

	DefaultTrenchLawHome = ".trenchsi"
	WorkspaceName       = "workspace"
)

// ResolveDefaultHome prefers the current home name, but reuses older homes
// when they already exist so renamed binaries continue to see prior installs.
func ResolveDefaultHome(userHome string) string {
	if userHome == "" {
		return ""
	}

	candidates := []string{
		DefaultTrenchLawHome,
		".trenchlaw",
		".jameclaw",
	}
	for _, name := range candidates {
		path := filepath.Join(userHome, name)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return filepath.Join(userHome, DefaultTrenchLawHome)
}
