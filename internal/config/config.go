package config

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	DefaultHost     = "http://localhost:11434"
	DefaultPlatform = "ollama"
	DefaultPort     = 11435
)

const (
	envAPIKey   = "OLLAMA_PROXY_API_KEY"
	envHost     = "OLLAMA_PROXY_HOST"
	envModel    = "OLLAMA_PROXY_MODEL"
	envPlatform = "OLLAMA_PROXY_PLATFORM"
	envPort     = "OLLAMA_PROXY_PORT"
)

type Config struct {
	APIKey   string
	Host     string
	Model    string
	Platform string
	Port     int
	Insecure bool
}

func Parse(args []string) (*Config, error) {
	fs := flag.NewFlagSet("ollama-proxy", flag.ContinueOnError)

	apiKey := fs.String("api-key", "", "API key required on incoming requests (env: OLLAMA_PROXY_API_KEY)")
	host := fs.String("host", "", "Upstream platform base URL (env: OLLAMA_PROXY_HOST)")
	model := fs.String("model", "", "Model to validate against the upstream (env: OLLAMA_PROXY_MODEL)")
	platform := fs.String("platform", "", "Upstream platform name (env: OLLAMA_PROXY_PLATFORM)")
	port := fs.Int("port", 0, "Port the proxy listens on (env: OLLAMA_PROXY_PORT)")
	insecure := fs.Bool("insecure", false, "Run without an API key without prompting")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	resolvedPort, err := resolvePort(*port)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		APIKey:   firstNonEmpty(*apiKey, os.Getenv(envAPIKey)),
		Host:     firstNonEmpty(*host, os.Getenv(envHost), DefaultHost),
		Model:    firstNonEmpty(*model, os.Getenv(envModel)),
		Platform: firstNonEmpty(*platform, os.Getenv(envPlatform), DefaultPlatform),
		Port:     resolvedPort,
		Insecure: *insecure,
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.Host == "" {
		return errors.New("host must not be empty")
	}
	u, err := url.Parse(c.Host)
	if err != nil {
		return fmt.Errorf("invalid host URL %q: %w", c.Host, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("host URL must use http or https scheme: %q", c.Host)
	}
	if u.Host == "" {
		return fmt.Errorf("host URL must include a host: %q", c.Host)
	}
	if c.Platform != DefaultPlatform {
		return fmt.Errorf("unsupported platform %q (only %q is supported)", c.Platform, DefaultPlatform)
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", c.Port)
	}
	return nil
}

func (c *Config) HostURL() (*url.URL, error) {
	return url.Parse(c.Host)
}

func (c *Config) SecurityEnabled() bool {
	return c.APIKey != ""
}

func resolvePort(flagPort int) (int, error) {
	if flagPort != 0 {
		return flagPort, nil
	}
	if env := strings.TrimSpace(os.Getenv(envPort)); env != "" {
		p, err := strconv.Atoi(env)
		if err != nil {
			return 0, fmt.Errorf("invalid %s value %q: %w", envPort, env, err)
		}
		return p, nil
	}
	return DefaultPort, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
