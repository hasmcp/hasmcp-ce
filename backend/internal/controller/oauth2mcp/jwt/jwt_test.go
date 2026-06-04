package jwt

import (
	"testing"
)

type mockConfig struct {
	secret string
}

func (m *mockConfig) Populate(key string, cfg any) error {
	if c, ok := cfg.(*jwtConfig); ok {
		c.Secret = m.secret
		return nil
	}
	return nil
}

func (m *mockConfig) Env() string          { return "" }
func (m *mockConfig) App() string          { return "" }
func (m *mockConfig) AppShortName() string { return "" }
func (m *mockConfig) Version() string      { return "" }

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		secret  string
		wantErr bool
	}{
		{
			name:    "too short",
			secret:  "short",
			wantErr: true,
		},
		{
			name:    "exactly 32",
			secret:  "12345678901234567890123456789012",
			wantErr: false,
		},
		{
			name:    "longer than 32",
			secret:  "1234567890123456789012345678901234567890",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(Params{
				Config: &mockConfig{secret: tt.secret},
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
