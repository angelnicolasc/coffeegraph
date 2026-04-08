package fsutil

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// CopyFSToDir copia recursivamente fsys (raíz root) hacia dst en disco.
func CopyFSToDir(fsys fs.FS, root, dst string) error {
	return fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		out := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(out, 0755)
		}
		rc, err := fsys.Open(path)
		if err != nil {
			return err
		}
		defer rc.Close()
		if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
			return err
		}
		w, err := os.Create(out)
		if err != nil {
			return err
		}
		defer w.Close()
		_, err = io.Copy(w, rc)
		return err
	})
}
