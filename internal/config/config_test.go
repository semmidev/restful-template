package config

import (
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "development environment defaults are allowed",
			config: Config{
				App:      App{Env: "development"},
				JWT:      JWT{Secret: "change-me-in-production-min-32-bytes!"},
				Asynqmon: Asynqmon{Password: "admin"},
				CORS:     CORS{AllowedOrigins: []string{"*"}},
			},
			wantErr: false,
		},
		{
			name: "production environment with defaults must fail",
			config: Config{
				App:      App{Env: "production"},
				JWT:      JWT{Secret: "change-me-in-production-min-32-bytes!"},
				Asynqmon: Asynqmon{Password: "admin"},
				CORS:     CORS{AllowedOrigins: []string{"https://app.example.com"}},
			},
			wantErr: true,
		},
		{
			name: "production environment with insecure JWT secret must fail",
			config: Config{
				App:      App{Env: "production"},
				JWT:      JWT{Secret: "short-secret"},
				Asynqmon: Asynqmon{Password: "secure-password"},
				CORS:     CORS{AllowedOrigins: []string{"https://app.example.com"}},
			},
			wantErr: true,
		},
		{
			name: "production environment with wildcard CORS must fail",
			config: Config{
				App:      App{Env: "production"},
				JWT:      JWT{Secret: "super-secure-secret-min-32-bytes-long-here!!!"},
				Asynqmon: Asynqmon{Password: "secure-password"},
				CORS:     CORS{AllowedOrigins: []string{"*"}},
			},
			wantErr: true,
		},
		{
			name: "production environment secure configuration must succeed",
			config: Config{
				App:      App{Env: "production"},
				JWT:      JWT{Secret: "super-secure-secret-min-32-bytes-long-here!!!"},
				Asynqmon: Asynqmon{Password: "secure-password"},
				CORS:     CORS{AllowedOrigins: []string{"https://app.example.com"}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
