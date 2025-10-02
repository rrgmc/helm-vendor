package file

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

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

func CopyFile(srcRoot, dstRoot *os.Root, src, dst string) error {
	sourceFile, err := srcRoot.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close() // Ensure the source file is closed

	destinationFile, err := dstRoot.Create(dst)
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

func Exists(root *os.Root, filePath string) bool {
	_, err := root.Stat(filePath)
	if err == nil {
		return true
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	return false
}

// GenerateUniqueFilename attempts to create a file with a unique name in the specified directory.
// It appends a counter to the base filename until a non-existent name is found and the file is created.
func GenerateUniqueFilename(root *os.Root, dir, baseName, extension string) (string, error) {
	for i := 0; i < 1000; i++ { // Limit retries to prevent infinite loops
		var filename string
		if i == 0 {
			filename = baseName + extension
		} else {
			filename = fmt.Sprintf("%s_%d%s", baseName, i, extension)
		}

		filePath := path.Join(dir, filename)

		_, err := root.Stat(filePath)
		if errors.Is(err, fs.ErrExist) {
			continue // File exists, try next iteration
		} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("failed to stat file %s: %w", filePath, err)
		}
		return filePath, nil // Success, return the path and the opened file
	}
	return "", fmt.Errorf("could not generate a unique filename after multiple attempts")
}
