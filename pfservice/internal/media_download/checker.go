package media

import (
	"os"
	"path/filepath"
)

// ImageExists checks if an image file exists on disk
func ImageExists(imagePath string) bool {
	if imagePath == "" {
		return false
	}

	fullPath := filepath.Join(MediaRoot, imagePath)
	_, err := os.Stat(fullPath)
	return err == nil
}

// GetFullImagePath returns the full filesystem path for a relative image path
func GetFullImagePath(imagePath string) string {
	return filepath.Join(MediaRoot, imagePath)
}

