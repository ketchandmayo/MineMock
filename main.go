package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"

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
	voicechatListenAddr := net.JoinHostPort(cfg.IP, strconv.Itoa(cfg.SimpleVoicechatPort))
	writeBanner()
	logServerConfig(cfg, voicechatListenAddr)

	loginCfg := server.LoginConfig{
		ErrorMessage:               cfg.ErrorMessage,
		ErrorDelay:                 cfg.ErrorDelay,
		ForceConnectionLostTitle:   cfg.ForceConnectionLostTitle,
		RealServerAddr:             cfg.RealServerAddr,
		IsWhitelisted:              cfg.IsLoginWhitelisted,
		SimpleVoicechatListenAddr:  voicechatListenAddr,
		SimpleVoicechatBackendAddr: cfg.RealServerVoicechatAddress(),
	}

	if err := server.Run(addr, statusCfg, loginCfg); err != nil {
		log.Println("Server error:", err)
		os.Exit(1)
	}
}

func logServerConfig(cfg config.Config, voicechatListenAddr string) {
	whitelist := make([]string, 0, len(cfg.LoginWhitelist))
	for username := range cfg.LoginWhitelist {
		whitelist = append(whitelist, username)
	}
	sort.Strings(whitelist)

	whitelistText := "<empty>"
	if len(whitelist) > 0 {
		whitelistText = strings.Join(whitelist, ", ")
	}

	realServerAddr := cfg.RealServerAddr
	if strings.TrimSpace(realServerAddr) == "" {
		realServerAddr = "<empty>"
	}

	voicechatBackendAddr := cfg.RealServerVoicechatAddress()
	if strings.TrimSpace(voicechatBackendAddr) == "" {
		voicechatBackendAddr = "<disabled>"
	}

	log.Printf(
		"Server configuration loaded:\n"+
			"  [network]\n"+
			"    listen_addr: %s\n"+
			"    ip: %s\n"+
			"    port: %s\n"+
			"  [status]\n"+
			"    motd: %q\n"+
			"    version_name: %q\n"+
			"    protocol: %d\n"+
			"    max_players: %d\n"+
			"    online_players: %d\n"+
			"  [login]\n"+
			"    error_delay: %s\n"+
			"    force_connection_lost_title: %t\n"+
			"    real_server_addr: %s\n"+
			"    whitelist_size: %d\n"+
			"    whitelist: %s\n"+
			"  [voicechat]\n"+
			"    listen_addr: %s\n"+
			"    backend_addr: %s",
		cfg.Address(),
		cfg.IP,
		cfg.Port,
		cfg.MOTD,
		cfg.VersionName,
		cfg.Protocol,
		cfg.MaxPlayers,
		cfg.OnlinePlayers,
		cfg.ErrorDelay,
		cfg.ForceConnectionLostTitle,
		realServerAddr,
		len(cfg.LoginWhitelist),
		whitelistText,
		voicechatListenAddr,
		voicechatBackendAddr,
	)
}

func writeBanner() {
	banner := `
	 ░  ░░░░  ░░        ░░   ░░░  ░░        ░░  ░░░░  ░░░      ░░░░      ░░░  ░░░░  ░
	 ▒   ▒▒   ▒▒▒▒▒  ▒▒▒▒▒    ▒▒  ▒▒  ▒▒▒▒▒▒▒▒   ▒▒   ▒▒  ▒▒▒▒  ▒▒  ▒▒▒▒  ▒▒  ▒▒▒  ▒▒
	 ▓        ▓▓▓▓▓  ▓▓▓▓▓  ▓  ▓  ▓▓      ▓▓▓▓        ▓▓  ▓▓▓▓  ▓▓  ▓▓▓▓▓▓▓▓     ▓▓▓▓
	 █  █  █  █████  █████  ██    ██  ████████  █  █  ██  ████  ██  ████  ██  ███  ██
	 █  ████  ██        ██  ███   ██        ██  ████  ███      ████      ███  ████  █

     by Ketch
	`

	fmt.Fprintln(os.Stdout, banner)
}
