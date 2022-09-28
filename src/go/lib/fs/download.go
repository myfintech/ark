package fs

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Download pulls a file from the internet
func Download(url, downloadDir string, perms os.FileMode) (downloadedFileName string, err error) {
	baseName := filepath.Base(url)
	downloadedFileName = filepath.Join(downloadDir, baseName)
	if err = os.MkdirAll(downloadDir, perms); err != nil {
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	fileHandle, err := os.OpenFile(downloadedFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perms)
	if err != nil {
		return
	}
	defer func() {
		_ = fileHandle.Close()
	}()

	if _, err = io.Copy(fileHandle, resp.Body); err != nil {
		return
	}

	return
}
