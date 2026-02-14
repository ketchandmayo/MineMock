package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	envIP                       = "IP"
	envPort                     = "PORT"
	envError                    = "ERROR"
	envErrorDelaySeconds        = "ERROR_DELAY_SECONDS"
	envForceConnectionLostTitle = "FORCE_CONNECTION_LOST_TITLE"
	envMOTD                     = "MOTD"
	envVersionName              = "VERSION_NAME"
	envProtocol                 = "PROTOCOL"
	envMaxPlayers               = "MAX_PLAYERS"
	envOnlinePlayers            = "ONLINE_PLAYERS"
)

const (
	defaultIP                  = "127.0.0.1"
	defaultPort                = "25565"
	defaultVersionName         = "1.20.1"
	defaultProtocol      int32 = 763
	defaultMaxPlayers          = 20
	defaultOnlinePlayers       = 7
)

const (
	defaultErrorMessage = "§r§7MineMock§r\\n\u00a7cServer is temporarily unavailable. Try again later."
	defaultMOTD         = "\u00a7c\u00a7oMine\u00a74\u00a7oMock\u00a7r\\n\u00a76Minecraft mock server on golang\u00a7r | \u00a7eWelcome\u263a"
)

type Config struct {
	IP                       string
	Port                     string
	ErrorMessage             string
	ErrorDelay               time.Duration
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

func FromEnv() Config {
	versionName := stringFromEnv(envVersionName, defaultVersionName)

	return Config{
		IP:                       stringFromEnv(envIP, defaultIP),
		Port:                     stringFromEnv(envPort, defaultPort),
		ErrorMessage:             decodedStringFromEnv(envError, defaultErrorMessage),
		ErrorDelay:               secondsDurationFromEnv(envErrorDelaySeconds, 0),
		ForceConnectionLostTitle: boolFromEnv(envForceConnectionLostTitle, false),
		MOTD:                     decodedStringFromEnv(envMOTD, defaultMOTD),
		VersionName:              versionName,
		Protocol:                 protocolFromEnv(versionName),
		MaxPlayers:               int32FromEnv(envMaxPlayers, defaultMaxPlayers),
		OnlinePlayers:            int32FromEnv(envOnlinePlayers, defaultOnlinePlayers),
	}
}

func secondsDurationFromEnv(key string, fallbackSeconds int64) time.Duration {
	parsed, ok := int64FromEnv(key)
	if !ok || parsed < 0 {
		return time.Duration(fallbackSeconds) * time.Second
	}

	return time.Duration(parsed) * time.Second
}

func decodedStringFromEnv(key string, fallback string) string {
	if value, ok := lookupNonEmptyEnv(key); ok {
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
	if value, ok := lookupNonEmptyEnv(key); ok {
		return value
	}

	return fallback
}

func int32FromEnv(key string, fallback int32) int32 {
	parsed, ok := int64FromEnv(key)
	if !ok {
		return fallback
	}

	return int32(parsed)
}

func int64FromEnv(key string) (int64, bool) {
	value, ok := lookupNonEmptyEnv(key)
	if !ok {
		return 0, false
	}

	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0, false
	}

	return parsed, true
}

func boolFromEnv(key string, fallback bool) bool {
	value, ok := lookupNonEmptyEnv(key)
	if !ok {
		return fallback
	}

	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}

	return parsed
}

func protocolFromEnv(versionName string) int32 {
	if parsed, ok := int32FromEnvValue(envProtocol); ok {
		return parsed
	}

	if protocol, ok := versionProtocolMap[strings.TrimSpace(versionName)]; ok {
		return protocol
	}

	return defaultProtocol
}

func int32FromEnvValue(key string) (int32, bool) {
	value, ok := int64FromEnv(key)
	if !ok {
		return 0, false
	}

	return int32(value), true
}

func lookupNonEmptyEnv(key string) (string, bool) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", false
	}

	if strings.TrimSpace(value) == "" {
		return "", false
	}

	return value, true
}

func (c Config) Address() string {
	return c.IP + ":" + c.Port
}
