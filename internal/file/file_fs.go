package file

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
)

func CopyFileFS(srcFS, dstFS *os.Root, src, dst string) error {
	sourceFile, err := srcFS.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close() // Ensure the source file is closed

	destinationFile, err := dstFS.Create(dst)
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

func ExistsFS(fsys *os.Root, filePath string) bool {
	_, err := fsys.Stat(filePath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// GenerateUniqueFilenameFS attempts to create a file with a unique name in the specified directory.
// It appends a counter to the base filename until a non-existent name is found and the file is created.
func GenerateUniqueFilenameFS(fsys *os.Root, dir, baseName, extension string) (string, error) {
	for i := 0; i < 1000; i++ { // Limit retries to prevent infinite loops
		var filename string
		if i == 0 {
			filename = baseName + extension
		} else {
			filename = fmt.Sprintf("%s_%d%s", baseName, i, extension)
		}

		filePath := path.Join(dir, filename)

		_, err := fsys.Stat(filePath)
		if errors.Is(err, fs.ErrExist) {
			continue // File exists, try next iteration
		} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("failed to stat file %s: %w", filePath, err)
		}
		return filePath, nil // Success, return the path and the opened file
	}
	return "", fmt.Errorf("could not generate a unique filename after multiple attempts")
}
