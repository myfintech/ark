package autoupdate

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"

	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/fs"

	"github.com/hashicorp/go-version"
)

var localVersionInfo *LocalVersionInfo

// LocalVersionInfo contains version info for the locally running binary
type LocalVersionInfo struct {
	Name                 string `json:"name"`
	Version              string `json:"version"`
	VersionCheckEndpoint string `json:"version_check_endpoint"`
}

// RemoteVersionInfo contains version info for the published remote binary
type RemoteVersionInfo struct {
	Version    string      `json:"version"`
	OSPackages []OSPackage `json:"os_packages"`
}

// OSPackage contains OS specific info for updating the correct version of the binary
type OSPackage struct {
	URL      string `json:"url"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Checksum string `json:"checksum"`
}

func Init(version *LocalVersionInfo) {
	localVersionInfo = version
}

func InitFromString(version string) error {
	versionBytes := []byte(version)

	if err := json.Unmarshal(versionBytes, &localVersionInfo); err != nil {
		return err
	}

	return nil
}

func CheckVersion() (bool, *RemoteVersionInfo, error) {
	remoteVersionInfo := new(RemoteVersionInfo)

	resp, err := http.Get(localVersionInfo.VersionCheckEndpoint)
	if err != nil {
		return false, remoteVersionInfo, errors.Wrapf(err, "GET %s failed", localVersionInfo.VersionCheckEndpoint)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, remoteVersionInfo, errors.Wrapf(err, "Body Read %s failed", localVersionInfo.VersionCheckEndpoint)
	}

	if err = json.Unmarshal(respBytes, remoteVersionInfo); err != nil {
		return false, remoteVersionInfo, errors.Wrapf(err, "failed to parse remove version info %s", localVersionInfo.VersionCheckEndpoint)
	}

	localVersion, err := version.NewVersion(localVersionInfo.Version)
	if err != nil {
		return false, remoteVersionInfo, errors.Wrapf(err, "failed to parse local version %s", localVersionInfo.Version)
	}

	remoteVersion, err := version.NewVersion(remoteVersionInfo.Version)
	if err != nil {
		return false, remoteVersionInfo, errors.Wrapf(err, "failed to parse remote version %s", remoteVersionInfo.Version)
	}

	if localVersion.LessThan(remoteVersion) {
		return true, remoteVersionInfo, nil
	}

	return false, remoteVersionInfo, nil
}

// DownloadBinary downloads a binary from a remote location and saves it to disk
func DownloadBinary(url string) (string, error) {
	client := new(http.Client)

	file, err := os.CreateTemp("", localVersionInfo.Name)
	if err != nil {
		return "", errors.Wrap(err, "failed to create temp file")
	}

	defer func() {
		_ = file.Close()
	}()

	request, err := http.NewRequest("GET", url, nil)
	request.Header.Add("Accept-Encoding", "gzip")
	if err != nil {
		return "", errors.Wrapf(err, "failed to create HTTP Request to %s", url)
	}

	resp, err := client.Do(request)
	if err != nil {
		return "", errors.Wrapf(err, "GET %s failed", url)
	}
	defer func() { _ = resp.Body.Close() }()

	var reader = resp.Body

	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return "", errors.Wrap(err, "failed to create gzip reader")
		}
	}

	defer func() { _ = reader.Close() }()

	_, err = io.Copy(file, reader)
	if err != nil {
		return file.Name(), errors.Wrapf(err, "failed to copy to %s", file.Name())
	}

	return file.Name(), nil
}

func SelectVersionByOS(packages []OSPackage) *OSPackage {
	for _, pkg := range packages {
		if pkg.OS == runtime.GOOS && pkg.Arch == runtime.GOARCH {
			return &pkg
		}
	}
	return nil
}

func Upgrade(binPath string) error {
	_, remoteVersionInfo, err := CheckVersion()
	if err != nil {
		return err
	}

	selected := SelectVersionByOS(remoteVersionInfo.OSPackages)

	if selected == nil {
		return errors.Errorf("no matching version for %s-%s", runtime.GOOS, runtime.GOARCH)
	}

	downloadPath, err := DownloadBinary(selected.URL)
	if err != nil {
		return errors.Wrapf(err, "failed to download %s", selected.URL)
	}

	checksum, err := fs.HashFile(downloadPath, sha256.New())
	if err != nil {
		return errors.Wrapf(err, "failed to hash %s", downloadPath)
	}

	downloadChecksum := hex.EncodeToString(checksum.Sum(nil))

	if downloadChecksum != selected.Checksum {
		return errors.Errorf("checksum mismatch: %s had %s, expected %s", downloadPath, downloadChecksum, selected.Checksum)
	}

	if err = os.Rename(downloadPath, binPath); err != nil {
		return errors.Wrapf(err, "failed to overwrite destination %s with %s", binPath, downloadPath)
	}

	if err = os.Chmod(binPath, 0711); err != nil {
		return errors.Wrapf(err, "failed to mark %s as executable after upgrade", binPath)
	}

	if err = os.RemoveAll(downloadPath); err != nil {
		return errors.Wrapf(err, "failed to cleanup %s", downloadPath)
	}

	return nil
}

func CurrentVersion() LocalVersionInfo {
	return *localVersionInfo
}
