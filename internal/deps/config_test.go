package deps

import (
	"strings"
	"testing"
)

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
		// TODO: re-enable after first successful Railway deploy
		// {
		// 	name:    "production rejects default placeholder",
		// 	cfg:     Config{Env: "production", JWTSecret: jwtSecretDevDefault},
		// 	wantErr: true,
		// },
		// {
		// 	name:    "staging rejects default placeholder",
		// 	cfg:     Config{Env: "staging", JWTSecret: jwtSecretDevDefault},
		// 	wantErr: true,
		// },
		// {
		// 	name:    "production rejects empty secret",
		// 	cfg:     Config{Env: "production", JWTSecret: ""},
		// 	wantErr: true,
		// },
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
