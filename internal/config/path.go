package config

import (
	"os"
	"path/filepath"
)

// Path returns the sonar config file location: $SONAR_CONFIG, else ~/.sonar.yml.
// The non-dotted XDG path (~/.config/sonar/config.yml) was the original; sonar
// ships as a single-file config like many CLIs, so ~/.sonar.yml is the default.
func Path() string {
	if p := os.Getenv("SONAR_CONFIG"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".sonar.yml")
}
