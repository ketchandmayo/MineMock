package config

import (
	"strings"
	"testing"
	"time"
)

func TestProtocolFromEnv_UsesVersionMapWhenProtocolMissing(t *testing.T) {
	t.Setenv("VERSION_NAME", "1.20.1")
	t.Setenv("PROTOCOL", "")

	cfg := FromEnv()
	if cfg.Protocol != 763 {
		t.Fatalf("expected protocol 763 for 1.20.1, got %d", cfg.Protocol)
	}
}

func TestProtocolFromEnv_ProtocolEnvOverridesVersion(t *testing.T) {
	t.Setenv("VERSION_NAME", "1.20.1")
	t.Setenv("PROTOCOL", "999")

	cfg := FromEnv()
	if cfg.Protocol != 999 {
		t.Fatalf("expected explicit protocol override 999, got %d", cfg.Protocol)
	}
}

func TestProtocolFromEnv_UnknownVersionFallback(t *testing.T) {
	t.Setenv("VERSION_NAME", "9.9.9")
	t.Setenv("PROTOCOL", "")

	cfg := FromEnv()
	if cfg.Protocol != 763 {
		t.Fatalf("expected fallback protocol 763, got %d", cfg.Protocol)
	}
}

func TestFromEnv_DecodesServerPropertiesFormatting(t *testing.T) {
	t.Setenv("MOTD", `\u00a7c\u00a7oMine\u00a74\u00a7oMock\u00a7r\n\u00a76Minecraft mock server on golang\u00a7r | \u00a7eWelcome\u263a`)
	t.Setenv("ERROR", `\u00a7cОшибка\u00a7r\n\u00a77Попробуйте позже`)

	cfg := FromEnv()

	if !strings.Contains(cfg.MOTD, "\n") {
		t.Fatalf("expected decoded MOTD to contain real newline, got %q", cfg.MOTD)
	}
	if strings.Contains(cfg.MOTD, `\n`) {
		t.Fatalf("expected decoded MOTD to not contain escaped newline sequence, got %q", cfg.MOTD)
	}
	if !strings.Contains(cfg.MOTD, "§") || !strings.Contains(cfg.MOTD, "☺") {
		t.Fatalf("expected decoded MOTD to include unicode formatting and symbol, got %q", cfg.MOTD)
	}

	if !strings.Contains(cfg.ErrorMessage, "\n") {
		t.Fatalf("expected decoded ERROR to contain real newline, got %q", cfg.ErrorMessage)
	}
	if strings.Contains(cfg.ErrorMessage, `\n`) {
		t.Fatalf("expected decoded ERROR to not contain escaped newline sequence, got %q", cfg.ErrorMessage)
	}
	if !strings.Contains(cfg.ErrorMessage, "§") {
		t.Fatalf("expected decoded ERROR to include section sign formatting, got %q", cfg.ErrorMessage)
	}
}

func TestFromEnv_ForceConnectionLostTitle(t *testing.T) {
	t.Setenv("FORCE_CONNECTION_LOST_TITLE", "true")

	cfg := FromEnv()
	if !cfg.ForceConnectionLostTitle {
		t.Fatal("expected ForceConnectionLostTitle to be true")
	}
}

func TestFromEnv_ErrorDelaySeconds(t *testing.T) {
	t.Setenv("ERROR_DELAY_SECONDS", "3")

	cfg := FromEnv()
	if cfg.ErrorDelay != 3*time.Second {
		t.Fatalf("expected ErrorDelay to be 3s, got %s", cfg.ErrorDelay)
	}
}

func TestFromEnv_ErrorDelaySecondsDefaultOnInvalid(t *testing.T) {
	t.Setenv("ERROR_DELAY_SECONDS", "-1")

	cfg := FromEnv()
	if cfg.ErrorDelay != 0 {
		t.Fatalf("expected ErrorDelay to fallback to 0s, got %s", cfg.ErrorDelay)
	}
}

func TestFromEnv_ProtocolTrimmedValue(t *testing.T) {
	t.Setenv("VERSION_NAME", "1.20.1")
	t.Setenv("PROTOCOL", " 768 ")

	cfg := FromEnv()
	if cfg.Protocol != 768 {
		t.Fatalf("expected explicit protocol override 768, got %d", cfg.Protocol)
	}
}

func TestFromEnv_BlankStringEnvFallsBack(t *testing.T) {
	t.Setenv("IP", "   ")
	t.Setenv("PORT", "")

	cfg := FromEnv()
	if cfg.IP != "127.0.0.1" {
		t.Fatalf("expected fallback IP 127.0.0.1, got %q", cfg.IP)
	}
	if cfg.Port != "25565" {
		t.Fatalf("expected fallback PORT 25565, got %q", cfg.Port)
	}
}

func TestFromEnv_LoginWhitelistParsing(t *testing.T) {
	t.Setenv("LOGIN_WHITELIST", "Steve, Alex ;  Notch")

	cfg := FromEnv()

	if len(cfg.LoginWhitelist) != 3 {
		t.Fatalf("expected 3 whitelist users, got %d", len(cfg.LoginWhitelist))
	}
	if !cfg.IsLoginWhitelisted("steve") {
		t.Fatal("expected steve to be whitelisted")
	}
	if !cfg.IsLoginWhitelisted("ALEX") {
		t.Fatal("expected ALEX to be whitelisted (case-insensitive)")
	}
	if cfg.IsLoginWhitelisted("Herobrine") {
		t.Fatal("expected Herobrine not to be whitelisted")
	}
}

func TestFromEnv_RealServerAddr(t *testing.T) {
	t.Setenv("REAL_SERVER_ADDR", "play.example.com:25565")

	cfg := FromEnv()

	if cfg.RealServerAddr != "play.example.com:25565" {
		t.Fatalf("unexpected RealServerAddr: %q", cfg.RealServerAddr)
	}
}
