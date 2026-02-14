package config

import (
	"os"
	"strconv"
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

// FromEnv собирает конфигурацию из переменных окружения.
func FromEnv() Config {
	protocol := int32FromEnv("PROTOCOL", 760)
	maxPlayers := int32FromEnv("MAX_PLAYERS", 20)
	onlinePlayers := int32FromEnv("ONLINE_PLAYERS", 0)

	return Config{
		IP:            os.Getenv("IP"),
		Port:          os.Getenv("PORT"),
		ErrorMessage:  os.Getenv("ERROR"),
		MOTD:          stringFromEnv("MOTD", "§6MineMock §7| §fДобро пожаловать!"),
		VersionName:   stringFromEnv("VERSION_NAME", "1.19.4"),
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

// Address возвращает адрес в формате host:port.
func (c Config) Address() string {
	return c.IP + ":" + c.Port
}
