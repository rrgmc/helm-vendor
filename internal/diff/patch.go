package diff

import (
	"bytes"
	"fmt"
	"iter"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
)

type Patcher struct {
	files []*gitdiff.File
}

func NewPatcher(diffBody string) (*Patcher, error) {
	files, _, err := gitdiff.Parse(bytes.NewBufferString(diffBody))
	if err != nil {
		return nil, fmt.Errorf("error parsing unified diff: %w", err)
	}

	return &Patcher{
		files: files,
	}, nil
}

func (p *Patcher) Files() iter.Seq[*gitdiff.File] {
	return func(yield func(*gitdiff.File) bool) {
		for _, file := range p.files {
			if !yield(file) {
				return
			}
		}
	}
}
