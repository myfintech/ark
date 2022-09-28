package jsonnetutils

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// ReadInput gets Jsonnet code from the given place (file, commandline, stdin).
// It also updates the given filename to <stdin> or <cmdline> if it wasn't a
// real filename.
func ReadInput(filenameIsCode bool, filename string) (input string, err error) {
	if filenameIsCode {
		input, err = filename, nil
		filename = "<cmdline>"
	} else if filename == "-" {
		var bytes []byte
		bytes, err = ioutil.ReadAll(os.Stdin)
		input = string(bytes)
		filename = "<stdin>"
	} else {
		var bytes []byte
		bytes, err = ioutil.ReadFile(filename)
		input = string(bytes)
	}
	return
}

// WriteOutputStream writes the output as a YAML stream.
func WriteOutputStream(output []string, dest *os.File) (err error) {
	for _, doc := range output {
		_, err := dest.WriteString("---\n")
		if err != nil {
			return err
		}
		_, err = dest.WriteString(doc)
		if err != nil {
			return err
		}
	}

	return nil
}

// BuildLibrary ...
func BuildLibrary(basePath string, baseLibraries ...[]string) (exportedLibraries []string) {
	for _, libraryPaths := range baseLibraries {
		for _, library := range libraryPaths {
			if !filepath.IsAbs(library) {
				library = filepath.Clean(filepath.Join(basePath, library))
			}
			exportedLibraries = append(exportedLibraries, library)
		}
	}
	return
}
