package auth

import "testing"

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		wantErr     bool
		checkLength bool
	}{
		{
			name:        "Valid password",
			password:    "testPassword123",
			wantErr:     false,
			checkLength: true,
		},
		{
			name:        "Empty password",
			password:    "",
			wantErr:     false,
			checkLength: true,
		},
		{
			name:        "Long password",
			password:    "verylongpasswordthatishardtoremember123!@#",
			wantErr:     false,
			checkLength: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkLength && len(hash) == 0 {
				t.Error("HashPassword() returned empty hash")
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "Valid password match",
			password: "correctPassword123",
			want:     true,
		},
		{
			name:     "Invalid password",
			password: "wrongPassword123",
			want:     false,
		},
		{
			name:     "Empty password",
			password: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create hash from the correct password for testing
			hash, err := HashPassword("correctPassword123")
			if err != nil {
				t.Fatalf("Failed to create hash for test: %v", err)
			}

			if got := ValidatePassword(tt.password, hash); got != tt.want {
				t.Errorf("ValidatePassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPasswordHashUniqueness(t *testing.T) {
	password := "samePassword123"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash1 == hash2 {
		t.Error("HashPassword() should generate different hashes for the same password")
	}
}
