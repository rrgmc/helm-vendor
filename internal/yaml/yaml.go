package yaml

import (
	"io"
	"os"

	"sigs.k8s.io/yaml"
)

func Decode(r io.Reader, data any) error {
	f, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return yaml.UnmarshalStrict(f, data)
}

func DecodeFile(fsys *os.Root, filename string, data any) error {
	f, err := fsys.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return Decode(f, data)
}
