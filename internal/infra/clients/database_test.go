package clients

import "testing"

func TestValidateDatabaseSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		source  string
		wantErr bool
	}{
		{name: "empty", source: "", wantErr: true},
		{name: "no scheme", source: "lynx:lynx@localhost:5432/lynx?sslmode=disable", wantErr: true},
		{name: "postgres", source: "postgres://lynx:lynx@localhost:5432/lynx?sslmode=disable", wantErr: false},
		{name: "postgresql", source: "postgresql://lynx:lynx@localhost:5432/lynx?sslmode=disable", wantErr: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateDatabaseSource(tc.source)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}
