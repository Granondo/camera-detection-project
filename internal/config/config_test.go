package config

import "testing"

func TestStorageLimitBytes(t *testing.T) {
	tests := []struct {
		name string
		gb   float64
		want int64
	}{
		{name: "zero", gb: 0, want: 0},
		{name: "negative", gb: -1, want: 0},
		{name: "whole_gb", gb: 70, want: 70 * 1024 * 1024 * 1024},
		{name: "fractional_gb", gb: 0.5, want: 512 * 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{StorageLimitGB: tt.gb}
			got := cfg.StorageLimitBytes()
			if got != tt.want {
				t.Fatalf("StorageLimitBytes() = %d, want %d", got, tt.want)
			}
		})
	}
}
