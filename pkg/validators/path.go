package validators

import (
	"path/filepath"
	"strings"
)

// ValidatePath ensures the path is safe and doesn't contain path traversal attempts
func ValidatePath(path string, baseDir string) (string, error) {
	// Clean the path to remove any ".." or "." components
	cleanPath := filepath.Clean(path)

	// Ensure the path doesn't contain any path traversal attempts
	if strings.Contains(cleanPath, "..") {
		return "", ErrInvalidPath
	}

	// If baseDir is provided, ensure the path is within the base directory
	if baseDir != "" {
		absPath, err := filepath.Abs(cleanPath)
		if err != nil {
			return "", err
		}

		absBaseDir, err := filepath.Abs(baseDir)
		if err != nil {
			return "", err
		}

		// Check if the path is within the base directory
		if !strings.HasPrefix(absPath, absBaseDir) {
			return "", ErrInvalidPath
		}
	}

	return cleanPath, nil
}

// ValidateFilePath ensures a file path is safe for file operations
func ValidateFilePath(filePath string, baseDir string) (string, error) {
	// Validate the path
	validPath, err := ValidatePath(filePath, baseDir)
	if err != nil {
		return "", err
	}

	// Additional checks for file paths
	if strings.Contains(validPath, "\x00") {
		return "", ErrInvalidPath
	}

	return validPath, nil
}

// ValidateDirectoryPath ensures a directory path is safe for directory operations
func ValidateDirectoryPath(dirPath string, baseDir string) (string, error) {
	// Validate the path
	validPath, err := ValidatePath(dirPath, baseDir)
	if err != nil {
		return "", err
	}

	// Ensure it's a directory path (ends with separator or is root)
	if !strings.HasSuffix(validPath, string(filepath.Separator)) && validPath != "." && validPath != "/" {
		validPath = validPath + string(filepath.Separator)
	}

	return validPath, nil
}
