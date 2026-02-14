package config

import (
	"testing"
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
