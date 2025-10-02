package diff

import (
	"bytes"
	"errors"
	"io/fs"
	"os"

	"github.com/aymanbagabas/go-udiff"
)

type Builder struct {
	buffer      bytes.Buffer
	diffEnabled bool
}

func NewBuilder(diffEnabled bool) *Builder {
	return &Builder{diffEnabled: diffEnabled}
}

func (b *Builder) Add(sourcePath, destPath string, sourceFS, destFS *os.Root, sourceFile, destFile string) error {
	if !b.diffEnabled {
		return nil
	}

	sourceFileData, err := sourceFS.ReadFile(sourceFile)
	if err != nil {
		return err
	}

	destFileData, err := destFS.ReadFile(destFile)
	if errors.Is(err, fs.ErrNotExist) {
		destFileData = []byte("")
	} else if err != nil {
		return err
	}

	// get a diff of the files
	edits := udiff.Bytes(sourceFileData, destFileData)
	if len(edits) > 0 {
		diffstr, err := udiff.ToUnified(sourcePath, destPath, string(sourceFileData), edits, udiff.DefaultContextLines)
		if err != nil {
			return err
		}
		if diffstr != "" {
			_, _ = b.buffer.WriteString(diffstr)
		}
	}

	return nil
}

func (b *Builder) IsEmpty() bool {
	return b.buffer.Len() == 0
}

func (b *Builder) Bytes() []byte {
	return b.buffer.Bytes()
}

func (b *Builder) String() string {
	return b.buffer.String()
}
