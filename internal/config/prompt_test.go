package config

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfirmInsecure(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"y\n", true},
		{"yes\n", true},
		{"Y\n", true},
		{"n\n", false},
		{"\n", false},
		{"", false},
		{"nonsense\n", false},
	}
	for _, tt := range tests {
		t.Run(strings.TrimSpace(tt.input), func(t *testing.T) {
			var out bytes.Buffer
			got, err := ConfirmInsecure(strings.NewReader(tt.input), &out)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
