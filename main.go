package main

import (
	"fmt"
	"os"

	"MineMock/internal/config"
	"MineMock/internal/server"
)

func main() {
	cfg := config.FromEnv()
	addr := cfg.Address()

	statusCfg := server.StatusConfig{
		MOTD:          cfg.MOTD,
		VersionName:   cfg.VersionName,
		Protocol:      cfg.Protocol,
		MaxPlayers:    cfg.MaxPlayers,
		OnlinePlayers: cfg.OnlinePlayers,
	}

	if err := server.Run(addr, cfg.ErrorMessage, statusCfg); err != nil {
		fmt.Println("Server error:", err)
		os.Exit(1)
	}
}
