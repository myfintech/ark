package fs

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// TarInjectorFunc is a function used to intercept and write last minute files to a tar file
type TarInjectorFunc func(tarWriter *tar.Writer) error

func tarWalker(tarWriter *tar.Writer, prefix string) filepath.WalkFunc {
	return func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		if prefix != "" {
			// trim the source directory from the file name before saving to the archive
			header.Name = TrimPrefix(file, prefix)
		}

		if writeErr := tarWriter.WriteHeader(header); writeErr != nil {
			return writeErr
		}

		fileHandle, openErr := os.Open(file)
		if openErr != nil {
			return openErr
		}
		defer fileHandle.Close()

		if _, copyErr := io.Copy(tarWriter, fileHandle); copyErr != nil {
			return copyErr
		}

		return nil
	}
}

// GzipTar recursively archives and compresses a given directory
func GzipTar(srcDir string, dest io.Writer) error {
	gzipWriter := gzip.NewWriter(dest)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	return filepath.Walk(srcDir, tarWalker(tarWriter, srcDir))
}

// GzipTarFiles takes an explicit list of files and generates a gzipped tar archive to the dest io.Writer
// If a prefix is supplied it will trim that prefix from the file names before writing to the archive
func GzipTarFiles(files []string, prefix string, dest io.Writer, injector TarInjectorFunc) error {
	if len(files) == 0 && injector == nil {
		return errors.New("files list cannot be empty while no injector was provided")
	}

	gzipWriter := gzip.NewWriter(dest)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	walker := tarWalker(tarWriter, prefix)
	for _, filePath := range files {
		info, err := os.Stat(filePath)
		if walkErr := walker(filePath, info, err); walkErr != nil {
			return walkErr
		}
	}

	if injector != nil {
		if err := injector(tarWriter); err != nil {
			return err
		}
	}

	return nil
}

// GzipUntar extracts a gzipped tar archive to the specified extraction dir
func GzipUntar(extractDir string, r io.Reader) error {
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	// there's no close method on this reader
	tarReader := tar.NewReader(gzipReader)

	for {
		header, tarErr := tarReader.Next()

		// reached the end of the tar file
		// nothing left to do
		if tarErr == io.EOF {
			return nil
		}

		// an unrecoverable error occurred
		if tarErr != nil {
			return errors.Wrap(tarErr, "failed to reach tar header")
		}

		// the tar file header was nil (it's a pointer)
		if header == nil {
			continue
		}

		if strings.Contains(header.Name, "..") {
			// if the path contains a relative reference, it could be a malicious file trying to make its way outside the scope of the extraction directory
			return errors.Errorf("file: '%s' contains a relative path that could be exploited", header.Name)
		}

		// construct the absolute archive extraction path
		absFilePath := filepath.Join(extractDir, header.Name)

		if header.Typeflag == tar.TypeDir {
			if mkdirErr := os.MkdirAll(absFilePath, 0755); mkdirErr != nil {
				return errors.Wrapf(mkdirErr, "failed to create directory %s", absFilePath)
			}
			continue
		}

		if header.Typeflag == tar.TypeReg {
			if _, err = os.Stat(filepath.Dir(absFilePath)); os.IsNotExist(err) {
				if mkdirErr := os.MkdirAll(filepath.Dir(absFilePath), 0755); mkdirErr != nil {
					return errors.Wrapf(mkdirErr, "failed to create directory %s", absFilePath)
				}
			}
			// reproduce the mode of the file in the tarball
			fileHandle, openErr := os.OpenFile(absFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if openErr != nil {
				return openErr
			}

			if _, copyErr := io.Copy(fileHandle, tarReader); copyErr != nil {
				return copyErr
			}

			// never defer close in a loop to avoid leaking open file handles
			fileHandle.Close()
		}
	}
}

// TarFile an in memory representation of a file for a tar archive
type TarFile struct {
	Name string
	Body []byte
	Mode int64
}

// Size returns the size of the TarFIle
func (t *TarFile) Size() int64 {
	return int64(len(t.Body))
}

// Header returns a tar file header
func (t *TarFile) Header() *tar.Header {
	return &tar.Header{
		Name:       t.Name,
		Size:       t.Size(),
		Mode:       t.Mode,
		ModTime:    time.Now(),
		AccessTime: time.Now(),
		ChangeTime: time.Now(),
	}
}

// Write writes this file to a tar writer
func (t *TarFile) Write(w *tar.Writer) error {
	if err := w.WriteHeader(t.Header()); err != nil {
		return err
	}
	if _, err := w.Write(t.Body); err != nil {
		return err
	}
	return nil
}

// InjectTarFiles injects in memory files into a tar writer
func InjectTarFiles(files []*TarFile) TarInjectorFunc {
	return func(w *tar.Writer) error {
		for _, file := range files {
			if err := file.Write(w); err != nil {
				return err
			}
		}
		return nil
	}
}
