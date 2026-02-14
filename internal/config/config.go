package config

import (
	"os"
	"strconv"
	"strings"
)

// Config содержит переменные окружения для запуска сервера.
type Config struct {
	IP            string
	Port          string
	ErrorMessage  string
	MOTD          string
	VersionName   string
	Protocol      int32
	MaxPlayers    int32
	OnlinePlayers int32
}

var versionProtocolMap = map[string]int32{
	"1.19.4": 762,
	"1.20":   763,
	"1.20.1": 763,
	"1.20.2": 764,
	"1.20.3": 765,
	"1.20.4": 765,
	"1.20.5": 766,
	"1.20.6": 766,
	"1.21":   767,
	"1.21.1": 767,
	"1.21.2": 768,
	"1.21.3": 768,
	"1.21.4": 769,
}

// FromEnv собирает конфигурацию из переменных окружения.
func FromEnv() Config {
	versionName := stringFromEnv("VERSION_NAME", "1.20.1")
	protocol := protocolFromEnv(versionName)
	maxPlayers := int32FromEnv("MAX_PLAYERS", 20)
	onlinePlayers := int32FromEnv("ONLINE_PLAYERS", 7)

	if maxPlayers <= 0 {
		maxPlayers = 20
	}
	if onlinePlayers < 0 {
		onlinePlayers = 0
	}
	if onlinePlayers > maxPlayers {
		onlinePlayers = maxPlayers
	}

	return Config{
		IP:            os.Getenv("IP"),
		Port:          os.Getenv("PORT"),
		ErrorMessage:  os.Getenv("ERROR"),
		MOTD:          stringFromEnv("MOTD", "§6MineMock §7| §fДобро пожаловать!"),
		VersionName:   versionName,
		Protocol:      protocol,
		MaxPlayers:    maxPlayers,
		OnlinePlayers: onlinePlayers,
	}
}

func stringFromEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func int32FromEnv(key string, fallback int32) int32 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return fallback
	}

	return int32(parsed)
}

func protocolFromEnv(versionName string) int32 {
	if value, ok := os.LookupEnv("PROTOCOL"); ok && value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err == nil {
			return int32(parsed)
		}
	}

	if protocol, ok := versionProtocolMap[strings.TrimSpace(versionName)]; ok {
		return protocol
	}

	return 763
}

// Address возвращает адрес в формате host:port.
func (c Config) Address() string {
	return c.IP + ":" + c.Port
}
