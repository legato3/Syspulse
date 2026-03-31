package pathutil

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
)

// NormalizeDir returns a clean absolute directory path.
func NormalizeDir(dir string) (string, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return "", fmt.Errorf("directory path cannot be empty")
	}

	abs, err := filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return "", fmt.Errorf("normalize directory path: %w", err)
	}
	return abs, nil
}

// JoinBaseFile safely joins a trusted root with a single basename child.
func JoinBaseFile(root, name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("file name cannot be empty")
	}
	if filepath.Base(name) != name || name == "." || name == ".." || strings.ContainsRune(name, 0) {
		return "", fmt.Errorf("unsafe file name %q", name)
	}

	joined := filepath.Join(root, name)
	rel, err := filepath.Rel(root, joined)
	if err != nil {
		return "", fmt.Errorf("resolve child path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes root")
	}
	return joined, nil
}

// LegacySafeFilename derives a filesystem-safe legacy filename.
func LegacySafeFilename(id, suffix string) (string, bool) {
	id = strings.TrimSpace(id)
	if id == "" || strings.ContainsRune(id, 0) {
		return "", false
	}

	var b strings.Builder
	for _, r := range id {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}

	name := strings.Trim(b.String(), "._")
	if name == "" || name == "." || name == ".." {
		return "", false
	}

	return name + suffix, true
}
