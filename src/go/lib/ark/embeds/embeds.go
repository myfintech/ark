package embeds

import (
	"embed"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Types embeds the @types dir with the binary at compile time
//go:embed .ark .ark/external_modules/.gitkeep .ark/native_modules/.gitkeep
var Types embed.FS

// Walk performs a recursive walk for a given directory
func Walk(dirFunc fs.WalkDirFunc) error {
	return fs.WalkDir(Types, ".ark", dirFunc)
}

// Unpack copy all the embedded files and directories from ark binary to .ark
func Unpack(baseDir string) error {
	return Walk(func(path string, d fs.DirEntry, err error) error {
		name := filepath.Join(baseDir, path)
		if err != nil {
			return nil
		}

		if d.IsDir() {
			return os.MkdirAll(name, 0755)
		}

		fBytes, err := Types.ReadFile(path)
		if err != nil {
			return err
		}
		return ioutil.WriteFile(name, fBytes, 0644)
	})
}
