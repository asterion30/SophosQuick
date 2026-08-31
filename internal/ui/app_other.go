//go:build !windows

package ui

import (
	"fmt"
	"sophosquick/internal/config"
	"sophosquick/internal/sophos"
)

type StubUI struct {
	cfg    *config.Config
	client *sophos.Client
}

func newPlatformUI(cfg *config.Config, client *sophos.Client) UI {
	return &StubUI{cfg: cfg, client: client}
}

func (s *StubUI) ShowAndRun() {
	fmt.Println("SophosQuick (Non-Windows mode)")
	fmt.Println("This application is designed for Windows. On Linux, use backend services or web client.")
}
