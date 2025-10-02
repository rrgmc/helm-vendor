package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close() // Ensure the source file is closed

	destinationFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destinationFile.Close() // Ensure the destination file is closed

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

func WithoutExt(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func NameExtFormat(fileName string) string {
	fne := WithoutExt(fileName)
	ext := strings.TrimPrefix(filepath.Ext(fileName), ".")
	if fne == "" {
		return fileName
	}
	if ext != "" {
		return fne + "_" + ext
	}
	return fne
}

func Exists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// GenerateUniqueFilename attempts to create a file with a unique name in the specified directory.
// It appends a counter to the base filename until a non-existent name is found and the file is created.
func GenerateUniqueFilename(dir, baseName, extension string) (string, error) {
	for i := 0; i < 1000; i++ { // Limit retries to prevent infinite loops
		var filename string
		if i == 0 {
			filename = baseName + extension
		} else {
			filename = fmt.Sprintf("%s_%d%s", baseName, i, extension)
		}

		filePath := filepath.Join(dir, filename)

		_, err := os.Stat(filePath)
		if os.IsExist(err) {
			continue // File exists, try next iteration
		} else if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("failed to stat file %s: %w", filePath, err)
		}
		return filePath, nil // Success, return the path and the opened file
	}
	return "", fmt.Errorf("could not generate a unique filename after multiple attempts")
}
