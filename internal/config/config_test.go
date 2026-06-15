package config

import (
	"strings"
	"testing"
)

func TestParseDefaults(t *testing.T) {
	t.Setenv(envAPIKey, "")
	t.Setenv(envHost, "")
	t.Setenv(envModel, "")
	t.Setenv(envPlatform, "")
	t.Setenv(envPort, "")

	cfg, err := Parse([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != DefaultHost {
		t.Errorf("host = %q, want %q", cfg.Host, DefaultHost)
	}
	if cfg.Platform != DefaultPlatform {
		t.Errorf("platform = %q, want %q", cfg.Platform, DefaultPlatform)
	}
	if cfg.Port != DefaultPort {
		t.Errorf("port = %d, want %d", cfg.Port, DefaultPort)
	}
	if cfg.SecurityEnabled() {
		t.Error("expected security disabled with no api key")
	}
}

func TestParseFlagsOverrideEnv(t *testing.T) {
	t.Setenv(envAPIKey, "env-key")
	t.Setenv(envHost, "http://env-host:1234")

	cfg, err := Parse([]string{"--api-key", "flag-key", "--host", "http://flag-host:9999", "--model", "llama3"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIKey != "flag-key" {
		t.Errorf("api key = %q, want flag-key", cfg.APIKey)
	}
	if cfg.Host != "http://flag-host:9999" {
		t.Errorf("host = %q", cfg.Host)
	}
	if cfg.Model != "llama3" {
		t.Errorf("model = %q", cfg.Model)
	}
	if !cfg.SecurityEnabled() {
		t.Error("expected security enabled")
	}
}

func TestParseEnvFallback(t *testing.T) {
	t.Setenv(envAPIKey, "env-key")
	t.Setenv(envHost, "")

	cfg, err := Parse([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIKey != "env-key" {
		t.Errorf("api key = %q, want env-key", cfg.APIKey)
	}
}

func TestParseInvalidHost(t *testing.T) {
	for _, h := range []string{"not-a-url", "ftp://host", "http://"} {
		t.Run(h, func(t *testing.T) {
			if _, err := Parse([]string{"--host", h}); err == nil {
				t.Errorf("expected error for host %q", h)
			}
		})
	}
}

func TestParseUnsupportedPlatform(t *testing.T) {
	if _, err := Parse([]string{"--platform", "vllm"}); err == nil {
		t.Fatal("expected error for unsupported platform")
	}
}

func TestParseInvalidPort(t *testing.T) {
	_, err := Parse([]string{"--port", "70000"})
	if err == nil || !strings.Contains(err.Error(), "port") {
		t.Fatalf("expected port error, got %v", err)
	}
}
