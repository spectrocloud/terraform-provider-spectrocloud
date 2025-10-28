package spectrocloud

import (
	"testing"
)

func TestParseClusterProfileImportID(t *testing.T) {
	tests := []struct {
		name           string
		importID       string
		wantContext    string
		wantIdentifier string
		wantErr        bool
	}{
		{
			name:           "valid UID with project context",
			importID:       "507f1f77bcf86cd799439011:project",
			wantContext:    "project",
			wantIdentifier: "507f1f77bcf86cd799439011",
			wantErr:        false,
		},
		{
			name:           "valid UID with tenant context",
			importID:       "507f1f77bcf86cd799439011:tenant",
			wantContext:    "tenant",
			wantIdentifier: "507f1f77bcf86cd799439011",
			wantErr:        false,
		},
		{
			name:           "valid UID with system context",
			importID:       "507f1f77bcf86cd799439011:system",
			wantContext:    "system",
			wantIdentifier: "507f1f77bcf86cd799439011",
			wantErr:        false,
		},
		{
			name:           "valid profile name with project context",
			importID:       "prod-k8s-profile:project",
			wantContext:    "project",
			wantIdentifier: "prod-k8s-profile",
			wantErr:        false,
		},
		{
			name:           "valid profile name with tenant context",
			importID:       "shared-profile:tenant",
			wantContext:    "tenant",
			wantIdentifier: "shared-profile",
			wantErr:        false,
		},
		{
			name:     "invalid format - missing context",
			importID: "profile-name",
			wantErr:  true,
		},
		{
			name:     "invalid format - too many parts",
			importID: "profile:name:extra",
			wantErr:  true,
		},
		{
			name:     "invalid context",
			importID: "profile-name:invalid",
			wantErr:  true,
		},
		{
			name:     "empty identifier",
			importID: ":project",
			wantErr:  true,
		},
		{
			name:     "empty context",
			importID: "profile-name:",
			wantErr:  true,
		},
		{
			name:     "completely empty",
			importID: "",
			wantErr:  true,
		},
		{
			name:           "UUID format",
			importID:       "550e8400-e29b-41d4-a716-446655440000:project",
			wantContext:    "project",
			wantIdentifier: "550e8400-e29b-41d4-a716-446655440000",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotContext, gotIdentifier, err := ParseClusterProfileImportID(tt.importID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseClusterProfileImportID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotContext != tt.wantContext {
					t.Errorf("ParseClusterProfileImportID() gotContext = %v, want %v", gotContext, tt.wantContext)
				}
				if gotIdentifier != tt.wantIdentifier {
					t.Errorf("ParseClusterProfileImportID() gotIdentifier = %v, want %v", gotIdentifier, tt.wantIdentifier)
				}
			}
		})
	}
}

func TestLooksLikeUID(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "UUID v4 format",
			s:    "550e8400-e29b-41d4-a716-446655440000",
			want: true,
		},
		{
			name: "Long UID with many hyphens",
			s:    "cluster-profile-507f-1f77-bcf8-6cd7",
			want: true,
		},
		{
			name: "MongoDB ObjectId style",
			s:    "507f1f77bcf86cd799439011",
			want: true,
		},
		{
			name: "Simple short name",
			s:    "prod",
			want: false,
		},
		{
			name: "Kebab case name with one hyphen",
			s:    "prod-k8s",
			want: false,
		},
		{
			name: "Kebab case name with two hyphens",
			s:    "prod-k8s-v2",
			want: false,
		},
		{
			name: "Short ID",
			s:    "abc123",
			want: false,
		},
		{
			name: "Empty string",
			s:    "",
			want: false,
		},
		{
			name: "Single character",
			s:    "a",
			want: false,
		},
		{
			name: "Medium length without hyphens",
			s:    "abcdefghij",
			want: false,
		},
		{
			name: "Long name with few hyphens",
			s:    "very-long-profile-name",
			want: false,
		},
		{
			name: "UID-like with exactly 3 hyphens and length > 10",
			s:    "abc-def-ghi-jkl",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := looksLikeUID(tt.s); got != tt.want {
				t.Errorf("looksLikeUID(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

// TestLooksLikeUID_EdgeCases tests edge cases for UID detection
func TestLooksLikeUID_EdgeCases(t *testing.T) {
	// Test the boundary conditions
	tests := []struct {
		name     string
		s        string
		expected bool
		reason   string
	}{
		{
			name:     "exactly 10 chars, no hyphens",
			s:        "abcdefghij",
			expected: false,
			reason:   "length check should fail (< 10 is false, == 10 might not trigger UID logic)",
		},
		{
			name:     "11 chars, no hyphens",
			s:        "abcdefghijk",
			expected: false,
			reason:   "no hyphens, so not UID-like",
		},
		{
			name:     "11 chars, 3 hyphens",
			s:        "ab-cd-ef-gh",
			expected: true,
			reason:   "length > 10 and 3+ hyphens indicates UID",
		},
		{
			name:     "UUID standard format",
			s:        "123e4567-e89b-12d3-a456-426614174000",
			expected: true,
			reason:   "standard UUID format with 4 hyphens and length 36",
		},
		{
			name:     "almost UUID but wrong length",
			s:        "123e4567-e89b-12d3-a456-42661417400",
			expected: true,
			reason:   "has 4 hyphens, length != 36 but > 10, so UID-like",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := looksLikeUID(tt.s)
			if got != tt.expected {
				t.Errorf("looksLikeUID(%q) = %v, expected %v. Reason: %s", 
					tt.s, got, tt.expected, tt.reason)
			}
		})
	}
}

