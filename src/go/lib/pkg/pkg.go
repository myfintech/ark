package pkg

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/myfintech/ark/src/go/lib/fs"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/log"
)

var info = PackageInfo{
	Version:     "0.0.0",
	Environment: "development",
	Hash:        "_",
}

func noCacheURL(rawURL string) string {
	parsedURL, _ := url.Parse(rawURL)
	query := parsedURL.Query()
	query.Add("_nocache", strconv.FormatInt(time.Now().Unix(), 10))
	parsedURL.RawQuery = query.Encode()
	return parsedURL.String()
}

// GlobalInfo returns a singleton of PackageInfo that should have package data injected with -ldflags
func GlobalInfo() PackageInfo {
	return info
}

// SetGlobalInfo a function that can modify the singleton PackageInfo
func SetGlobalInfo(latest PackageInfo) {
	info = latest
}

// PackageInfo a data container struct intended to be used with -ldflags and injecting package version data
type PackageInfo struct {
	Version           string `json:"version"`
	Environment       string `json:"environment"`
	Hash              string `json:"hash"`
	RemoteVersionURL  string `json:"remote_version_url"`
	LatestDownloadURL string `json:"latest_download_url"`
}

// ComputeHash modifies the current package with the hash of the currently running executable
func (current *PackageInfo) ComputeHash() error {
	path, err := os.Executable()
	if err != nil {
		return errors.Wrap(err, "failed to locate path to current executable")
	}

	hash, err := fs.HashFile(path, nil)
	if err != nil {
		return errors.Wrap(err, "failed to compute hash of current executable")
	}

	current.Hash = hex.EncodeToString(hash.Sum(nil))
	return nil
}

// ParseVersion parses a semantic version string
func (current PackageInfo) ParseVersion() (*version.Version, error) {
	return version.NewSemver(current.Version)
}

// Validate checks if the package has the necessary information to attempt a version check
func (current PackageInfo) Validate() bool {
	return current.LatestDownloadURL != "" && current.RemoteVersionURL != ""
}

// CheckVersion makes a call to the cdn to check the remote version against the current version
// If current is less that latest check version returns true
func (current PackageInfo) CheckVersion() (bool, PackageInfo, error) {
	latest := PackageInfo{}
	resp, err := http.Get(noCacheURL(current.RemoteVersionURL))
	if err != nil {
		return false, latest, err
	}
	log.Debug(resp.Request.URL)

	if resp.StatusCode != http.StatusOK {
		return false, latest, err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, latest, err
	}

	if err = json.Unmarshal(respBytes, &latest); err != nil {
		return false, latest, err
	}

	curr, err := current.ParseVersion()
	if err != nil {
		return false, latest, err
	}
	next, err := version.NewVersion(latest.Version)
	if err != nil {
		return false, latest, err
	}

	return curr.LessThan(next), latest, nil
}

// DownloadLatest downloads latest from the domain CDN
func (current PackageInfo) DownloadLatest() (string, error) {
	return fs.Download(current.LatestDownloadURL, "/tmp/go/pkg/upgrade", 0755)
}

// Upgrade downloads the latest version and overwrites the current version;
// if the upgrade fails, the binary is rolled back to the version that was already on disk;
// if the rollback fails, the function will panic
func Upgrade(latest PackageInfo) (err error) {
	// TODO: os.Executable may be a symlink (create a method to handle this case; see docs: https://golang.org/pkg/os/#Executable)
	currentBinary, err := os.Executable()
	if err != nil {
		return
	}

	// Get a copy of the current binary in case a rollback is needed
	backupOfCurrent := fmt.Sprintf("%s.%s", currentBinary, "bak")
	if err = fs.Copy(currentBinary, backupOfCurrent); err != nil {
		return err
	}
	defer func() { _ = os.Remove(backupOfCurrent) }()

	// Download the latest version
	updateFile, err := info.DownloadLatest()
	defer func() { _ = os.Remove(updateFile) }()
	if err != nil {
		return
	}

	// Compare the file contents to the package info's listed hash
	_, err = fs.CompareFileHash(updateFile, latest.Hash)
	if err != nil {
		return
	}

	// Set the execute bit on the new binary before moving it into place
	if err = os.Chmod(updateFile, 0755); err != nil {
		return
	}

	// Remove the current binary and defer a rollback operation if the binary replacement fails
	if err = os.Remove(currentBinary); err != nil {
		return
	}
	defer func() {
		if err != nil {
			panic(rollback(backupOfCurrent, currentBinary))
		}
	}()

	// Replace the binary with the latest version
	if err = fs.Copy(updateFile, currentBinary); err != nil {
		return
	}

	return err
}

// UpgradeCallback designates the format for callback functions within the pkg library
type UpgradeCallback func(current PackageInfo, latest PackageInfo) error

// VersionCheckHook prompts user for upgrade and upgrades the package in place if accepted
func VersionCheckHook(disabled bool, upgradeCallback, upgradeNotRequired UpgradeCallback) error {
	if disabled {
		return nil
	}

	if !GlobalInfo().Validate() {
		log.Debug("Skipping package version check. Missing remote package details.")
		return nil
	}

	upgradeAvailable, latest, err := GlobalInfo().CheckVersion()
	if err != nil {
		return errors.Wrap(err, "failed to check remote version")
	}

	if upgradeAvailable {
		if callbackErr := upgradeCallback(GlobalInfo(), latest); callbackErr != nil {
			return callbackErr
		}
	} else if upgradeNotRequired != nil {
		return upgradeNotRequired(GlobalInfo(), latest)
	}
	return nil
}

// rollback copies a backup of the src back to the destination
func rollback(src, dst string) error {
	return fs.Copy(src, dst)
}
