package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"MineMock/internal/config"
	"MineMock/internal/server"
)

func main() {
	cfg := config.FromEnv()

	cleanupLogger, err := setupLogger(cfg.LogFile)
	if err != nil {
		fmt.Println("Logger setup error:", err)
		os.Exit(1)
	}
	defer cleanupLogger()

	addr := cfg.Address()

	log.Printf("Server config: ip=%q port=%q log_file=%q error_delay=%s force_connection_lost_title=%t version=%q protocol=%d motd=%q max_players=%d online_players=%d", cfg.IP, cfg.Port, cfg.LogFile, cfg.ErrorDelay, cfg.ForceConnectionLostTitle, cfg.VersionName, cfg.Protocol, cfg.MOTD, cfg.MaxPlayers, cfg.OnlinePlayers)

	statusCfg := server.StatusConfig{
		MOTD:          cfg.MOTD,
		VersionName:   cfg.VersionName,
		Protocol:      cfg.Protocol,
		MaxPlayers:    cfg.MaxPlayers,
		OnlinePlayers: cfg.OnlinePlayers,
	}

	if err := server.Run(addr, cfg.ErrorMessage, cfg.ErrorDelay, cfg.ForceConnectionLostTitle, statusCfg); err != nil {
		fmt.Println("Server error:", err)
		os.Exit(1)
	}
}

func setupLogger(logFile string) (func(), error) {
	if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	log.SetOutput(io.MultiWriter(os.Stdout, file))
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	return func() {
		_ = file.Close()
	}, nil
}
