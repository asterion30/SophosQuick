package ui

import (
	"sophosquick/internal/config"
	"sophosquick/internal/sophos"
)

// UI represents the graphical interface contract for SophosQuick.
type UI interface {
	ShowAndRun()
}

// New creates the platform-specific UI instance.
func New(cfg *config.Config, client *sophos.Client) UI {
	return newPlatformUI(cfg, client)
}
