package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"MineMock/internal/config"
	"MineMock/internal/server"
)

func main() {
	logFile, err := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Println("Failed to open server.log:", err)
		os.Exit(1)
	}
	defer logFile.Close()

	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	cfg := config.FromEnv()
	addr := cfg.Address()

	statusCfg := server.StatusConfig{
		MOTD:          cfg.MOTD,
		VersionName:   cfg.VersionName,
		Protocol:      cfg.Protocol,
		MaxPlayers:    cfg.MaxPlayers,
		OnlinePlayers: cfg.OnlinePlayers,
	}

	log.Printf(
		"Server configuration loaded: ip=%s port=%s motd=%q version=%s protocol=%d max_players=%d online_players=%d error_delay=%s force_connection_lost_title=%t real_server_addr=%q whitelist_size=%d",
		cfg.IP,
		cfg.Port,
		cfg.MOTD,
		cfg.VersionName,
		cfg.Protocol,
		cfg.MaxPlayers,
		cfg.OnlinePlayers,
		cfg.ErrorDelay,
		cfg.ForceConnectionLostTitle,
		cfg.RealServerAddr,
		len(cfg.LoginWhitelist),
	)

	loginCfg := server.LoginConfig{
		ErrorMessage:             cfg.ErrorMessage,
		ErrorDelay:               cfg.ErrorDelay,
		ForceConnectionLostTitle: cfg.ForceConnectionLostTitle,
		RealServerAddr:           cfg.RealServerAddr,
		IsWhitelisted:            cfg.IsLoginWhitelisted,
	}

	if err := server.Run(addr, statusCfg, loginCfg); err != nil {
		log.Println("Server error:", err)
		os.Exit(1)
	}
}
