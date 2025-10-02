package file

import (
	"github.com/bmatcuk/doublestar/v4"
)

type Filter struct {
	Ignore []string
}

func IterFilter(iter Iter, filter Filter) Iter {
	return func(yield func(Info, error) bool) {
	fileloop:
		for fi, err := range iter {
			if err != nil {
				yield(Info{}, err)
				return
			}
			for _, opf := range filter.Ignore {
				if match, matchErr := doublestar.Match(opf, fi.Path); match {
					continue fileloop
				} else if matchErr != nil {
					yield(Info{}, matchErr)
					return
				}
			}
			if !yield(fi, nil) {
				return
			}
		}
	}
}
