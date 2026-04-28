package deps

import (
	"strings"
	"testing"
)

func TestConfig_RefreshPepper(t *testing.T) {
	t.Parallel()
	t.Run("dedicated pepper when set", func(t *testing.T) {
		t.Parallel()
		c := &Config{RefreshTokenPepper: "only-refresh", JWTSecret: "jwt"}
		if c.RefreshPepper() != "only-refresh" {
			t.Fatalf("expected dedicated pepper")
		}
	})
	t.Run("derived when empty", func(t *testing.T) {
		t.Parallel()
		c := &Config{JWTSecret: "abc"}
		if c.RefreshPepper() != "abc:ralts-refresh" {
			t.Fatalf("unexpected: %q", c.RefreshPepper())
		}
	})
	t.Run("nil config", func(t *testing.T) {
		t.Parallel()
		var c *Config
		if c.RefreshPepper() != "" {
			t.Fatalf("expected empty")
		}
	})
}

func TestValidateConfig_JWTSecret(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "development allows default placeholder",
			cfg:     Config{Env: appEnvDevelopment, JWTSecret: jwtSecretDevDefault},
			wantErr: false,
		},
		{
			name:    "test allows default placeholder",
			cfg:     Config{Env: appEnvTest, JWTSecret: jwtSecretDevDefault},
			wantErr: false,
		},
		{
			name:    "production rejects default placeholder",
			cfg:     Config{Env: "production", JWTSecret: jwtSecretDevDefault},
			wantErr: true,
		},
		{
			name:    "staging rejects default placeholder",
			cfg:     Config{Env: "staging", JWTSecret: jwtSecretDevDefault},
			wantErr: true,
		},
		{
			name:    "production rejects empty secret",
			cfg:     Config{Env: "production", JWTSecret: ""},
			wantErr: true,
		},
		{
			name:    "production accepts custom secret",
			cfg:     Config{Env: "production", JWTSecret: "a-strong-enough-secret-for-jwt"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateConfig(&tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), "JWT_SECRET") {
					t.Fatalf("expected JWT_SECRET in error, got: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
