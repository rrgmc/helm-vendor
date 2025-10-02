package file

import (
	"errors"
	"io/fs"
	"iter"
	"path/filepath"
)

type Info struct {
	Path string
	// FullPath string
	Entry fs.DirEntry
}

type Iter = iter.Seq2[Info, error]

func IterDir(fsys fs.FS, rootPath string) Iter {
	errEnd := errors.New("end")
	return func(yield func(Info, error) bool) {
		err := fs.WalkDir(fsys, rootPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if path == "" || path == "." {
				return nil
			}
			fi := Info{
				Path: filepath.ToSlash(path),
				// FullPath: filepath.Join(rootPath, path),
				Entry: d,
			}
			if !yield(fi, nil) {
				return errEnd
			}
			return nil
		})
		if err != nil && !errors.Is(err, errEnd) {
			yield(Info{}, err)
		}
	}
}
