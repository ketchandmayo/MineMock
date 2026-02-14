package config

import (
	"os"
	"strconv"
	"strings"
)

// Config содержит переменные окружения для запуска сервера.
type Config struct {
	IP                       string
	Port                     string
	ErrorMessage             string
	ForceConnectionLostTitle bool
	MOTD                     string
	VersionName              string
	Protocol                 int32
	MaxPlayers               int32
	OnlinePlayers            int32
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

	return Config{
		IP:                       os.Getenv("IP"),
		Port:                     os.Getenv("PORT"),
		ErrorMessage:             serverPropertiesStringFromEnv("ERROR", "\u00a7cServer is temporarily unavailable. Try again later.\u00a7r\\n\u00a77MineMock\u00a7r"),
		ForceConnectionLostTitle: boolFromEnv("FORCE_CONNECTION_LOST_TITLE", false),
		MOTD:                     serverPropertiesStringFromEnv("MOTD", "\u00a7c\u00a7oMine\u00a74\u00a7oMock\u00a7r\\n\u00a76Minecraft mock server on golang\u00a7r | \u00a7eWelcome\u263a"),
		VersionName:              versionName,
		Protocol:                 protocolFromEnv(versionName),
		MaxPlayers:               int32FromEnv("MAX_PLAYERS", 20),
		OnlinePlayers:            int32FromEnv("ONLINE_PLAYERS", 7),
	}
}

func serverPropertiesStringFromEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return decodeServerPropertiesEscapes(value)
	}

	return decodeServerPropertiesEscapes(fallback)
}

func decodeServerPropertiesEscapes(input string) string {
	decoded, err := strconv.Unquote(`"` + strings.ReplaceAll(input, `"`, `\\"`) + `"`)
	if err != nil {
		return input
	}

	return decoded
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

func boolFromEnv(key string, fallback bool) bool {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}

	return parsed
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
