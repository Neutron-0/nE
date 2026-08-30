package scanner

import (
	"regexp"
	"strings"

	"ne/internal/domain"
	"golang.org/x/text/unicode/norm"
)

var (
	featRegex   = regexp.MustCompile(`(?i)\s+(?:feat\.?|ft\.?|featuring)\s+`)
	vsRegex     = regexp.MustCompile(`(?i)\s+(?:vs\.?|versus)\s+`)
	multiSpaces = regexp.MustCompile(`\s+`)
)

// NormalizeString trims whitespace, collapses multiple spaces, and converts to Unicode NFC.
func NormalizeString(s string) string {
	s = strings.TrimSpace(s)
	s = multiSpaces.ReplaceAllString(s, " ")
	return norm.NFC.String(s)
}

// GenerateSortName formats names for alphabetical sorting (e.g., "The Beatles" -> "Beatles, The").
func GenerateSortName(name string, cleanEnglishArticles bool) string {
	name = NormalizeString(name)
	if !cleanEnglishArticles || len(name) < 4 {
		return name
	}

	lower := strings.ToLower(name)
	if strings.HasPrefix(lower, "the ") {
		return strings.TrimSpace(name[4:]) + ", The"
	}
	if strings.HasPrefix(lower, "a ") {
		return strings.TrimSpace(name[2:]) + ", A"
	}
	if strings.HasPrefix(lower, "an ") {
		return strings.TrimSpace(name[3:]) + ", An"
	}

	return name
}

// SplitArtists parses composite artist strings (e.g. "Daft Punk feat. Pharrell Williams / Nile Rodgers") into structured roles.
func SplitArtists(rawArtist string) []domain.ParsedArtist {
	rawArtist = NormalizeString(rawArtist)
	if rawArtist == "" {
		return []domain.ParsedArtist{{Name: "Unknown Artist", Role: domain.RolePrimary}}
	}

	var results []domain.ParsedArtist

	// Split by featuring first
	featParts := featRegex.Split(rawArtist, -1)
	primaryPart := featParts[0]

	// Split primary by standard delimiters (; , / &)
	primaryNames := splitDelimiters(primaryPart)
	for _, name := range primaryNames {
		name = NormalizeString(name)
		if name != "" {
			results = append(results, domain.ParsedArtist{Name: name, Role: domain.RolePrimary})
		}
	}

	// Any parts after feat. are featured artists
	if len(featParts) > 1 {
		for _, feat := range featParts[1:] {
			featuredNames := splitDelimiters(feat)
			for _, name := range featuredNames {
				name = NormalizeString(name)
				if name != "" {
					results = append(results, domain.ParsedArtist{Name: name, Role: domain.RoleFeatured})
				}
			}
		}
	}

	if len(results) == 0 {
		return []domain.ParsedArtist{{Name: rawArtist, Role: domain.RolePrimary}}
	}

	return results
}

func splitDelimiters(s string) []string {
	s = strings.ReplaceAll(s, " / ", ";")
	s = strings.ReplaceAll(s, " & ", ";")
	s = strings.ReplaceAll(s, " and ", ";")
	s = strings.ReplaceAll(s, " | ", ";")
	parts := strings.Split(s, ";")
	var cleaned []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			cleaned = append(cleaned, p)
		}
	}
	return cleaned
}
