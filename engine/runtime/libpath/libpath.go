// Package libpath provides utilities for identifying blessed library paths.
package libpath

import (
	"path/filepath"
	"strings"
)

// IsBlessedLibPath checks if a file path is in a blessed lib directory.
// Blessed directories are paths containing /lib/ as a directory component
// (not as part of another directory name like /stdlib/).
func IsBlessedLibPath(filePath string) bool {
	normalized := filepath.ToSlash(filePath)

	// Split into components and look for "lib" as a directory
	parts := strings.Split(normalized, "/")
	for i, part := range parts {
		if part == "lib" {
			// Found lib directory - it's blessed
			// Optionally check if preceded by "contrib"
			return true
		}
		if part == "contrib" && i+1 < len(parts) && parts[i+1] == "lib" {
			return true
		}
	}
	return false
}
