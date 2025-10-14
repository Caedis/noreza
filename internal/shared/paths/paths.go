package paths

import (
	"os"
	"path/filepath"
)

func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "noreza")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "noreza")
}

func DeviceDir(serial string) string {
	return filepath.Join(ConfigDir(), "devices", serial)
}

func ProfilesDir(serial string) string {
	return filepath.Join(DeviceDir(serial), "profiles")
}
