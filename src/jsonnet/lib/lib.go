//go:generate go run ../../go/tools/pack/main.go ./ ./generated.go jsonnetlib

package jsonnetlib

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/myfintech/ark/src/go/lib/fs"
	"github.com/myfintech/ark/src/go/lib/log"
)

const (
	// LibPath is the path to the ark jsonnet library
	LibPath = "/ark/lib/jsonnet"
)

// Install will extract the in memory archive to the lib path
func Install(installDir string) error {
	if installDir == "" {
		installDir = LibPath
	}

	var checksumPath = filepath.Join(installDir, "checksum")
	_, err := os.Stat(checksumPath)
	switch {
	case os.IsNotExist(err):
		break
	default:
		checksumBytes, readErr := ioutil.ReadFile(checksumPath)
		if readErr != nil {
			return readErr
		}

		if strings.Compare(string(checksumBytes), Hash) == 0 {
			return nil
		}
	}

	// clean the existing install path
	if err = os.RemoveAll(installDir); err != nil {
		return err
	}

	// recreate the initial install dir
	if err = os.MkdirAll(installDir, 0755); err != nil {
		return err
	}

	if err = fs.GzipUntar(installDir, Archive); err != nil {
		return err
	}

	if err = ioutil.WriteFile(checksumPath, []byte(Hash), 0644); err != nil {
		return err
	}

	log.Debugf("Successfully configured the ark jsonnet lib %s", installDir)

	return nil
}
