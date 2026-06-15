package upstream

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

type Platform struct {
	client  *Client
	command string
	args    []string
}

func NewOllamaPlatform(client *Client) *Platform {
	return &Platform{
		client:  client,
		command: "ollama",
		args:    []string{"serve"},
	}
}

func (p *Platform) Running(ctx context.Context) bool {
	return p.client.Ping(ctx) == nil
}

func (p *Platform) Start() error {
	if _, err := exec.LookPath(p.command); err != nil {
		return fmt.Errorf("%q not found in PATH: %w", p.command, err)
	}
	cmd := exec.Command(p.command, p.args...)
	configureDetached(cmd)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %q: %w", p.command, err)
	}
	return cmd.Process.Release()
}

func (p *Platform) WaitReady(ctx context.Context, timeout, interval time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if p.Running(ctx) {
			return nil
		}
		if time.Now().After(deadline) {
			return errors.New("timed out waiting for platform to become ready")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}

func (p *Platform) EnsureRunning(ctx context.Context, timeout, interval time.Duration) (bool, error) {
	if p.Running(ctx) {
		return false, nil
	}
	if err := p.Start(); err != nil {
		return false, err
	}
	if err := p.WaitReady(ctx, timeout, interval); err != nil {
		return true, err
	}
	return true, nil
}
