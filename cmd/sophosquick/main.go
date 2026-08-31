package main

import (
	"log"

	"sophosquick/internal/config"
	"sophosquick/internal/sophos"
	"sophosquick/internal/ui"
)

func main() {
	// 1. Load application configuration and preferences
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Warning: failed to load config file: %v (using defaults)", err)
		cfg = config.DefaultConfig()
	}

	// 2. Initialize Sophos Connect client interface
	client := sophos.NewClient(cfg.SccliPath)

	// 3. Instantiate and run modern Dark Slate UI
	gui := ui.New(cfg, client)
	gui.ShowAndRun()
}
