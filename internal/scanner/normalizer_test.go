package scanner_test

import (
	"testing"

	"ne/internal/domain"
	"ne/internal/scanner"
)

func TestGenerateSortName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"The Beatles", "Beatles, The"},
		{"The Rolling Stones", "Rolling Stones, The"},
		{"A Tribe Called Quest", "Tribe Called Quest, A"},
		{"Anathema", "Anathema"},
		{"An Autumn Breeze", "Autumn Breeze, An"},
		{"Radiohead", "Radiohead"},
	}

	for _, tt := range tests {
		result := scanner.GenerateSortName(tt.input, true)
		if result != tt.expected {
			t.Errorf("GenerateSortName(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSplitArtists(t *testing.T) {
	tests := []struct {
		input    string
		expected []domain.ParsedArtist
	}{
		{
			"Daft Punk feat. Pharrell Williams",
			[]domain.ParsedArtist{
				{Name: "Daft Punk", Role: domain.RolePrimary},
				{Name: "Pharrell Williams", Role: domain.RoleFeatured},
			},
		},
		{
			"Calvin Harris / Dua Lipa",
			[]domain.ParsedArtist{
				{Name: "Calvin Harris", Role: domain.RolePrimary},
				{Name: "Dua Lipa", Role: domain.RolePrimary},
			},
		},
		{
			"Single Artist",
			[]domain.ParsedArtist{
				{Name: "Single Artist", Role: domain.RolePrimary},
			},
		},
	}

	for _, tt := range tests {
		results := scanner.SplitArtists(tt.input)
		if len(results) != len(tt.expected) {
			t.Fatalf("SplitArtists(%q) returned %d artists, want %d", tt.input, len(results), len(tt.expected))
		}
		for i, r := range results {
			if r.Name != tt.expected[i].Name || r.Role != tt.expected[i].Role {
				t.Errorf("SplitArtists(%q)[%d] = %+v; want %+v", tt.input, i, r, tt.expected[i])
			}
		}
	}
}
