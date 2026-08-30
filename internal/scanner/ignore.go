package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type IgnoreMatcher struct {
	patterns []string
}

// NewIgnoreMatcher loads ignore patterns from an optional .neignore file.
func NewIgnoreMatcher(dir string, ignoreFilename string) *IgnoreMatcher {
	if ignoreFilename == "" {
		ignoreFilename = ".neignore"
	}

	matcher := &IgnoreMatcher{
		patterns: []string{
			".git", ".DS_Store", "Thumbs.db", "desktop.ini", "@eaDir",
			"*.tmp", "*.bak", "*.part",
		},
	}

	ignorePath := filepath.Join(dir, ignoreFilename)
	file, err := os.Open(ignorePath)
	if err != nil {
		return matcher
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		matcher.patterns = append(matcher.patterns, line)
	}

	return matcher
}

// ShouldIgnore tests if a filename or path matches any ignore patterns.
func (m *IgnoreMatcher) ShouldIgnore(path string) bool {
	base := filepath.Base(path)
	for _, pattern := range m.patterns {
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
		if strings.Contains(path, "/"+pattern+"/") || strings.Contains(path, "\\"+pattern+"\\") {
			return true
		}
	}
	return false
}
