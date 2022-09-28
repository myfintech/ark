package xdgbase

import (
	"fmt"
	"os"
	"strings"

	"github.com/myfintech/ark/src/go/lib/fs"
	"github.com/myfintech/ark/src/go/lib/utils"
)

// Suffix should be one of the predefined suffixes in this package
//
//	DataSuffix
//	ConfigSuffix
//	CacheSuffix
type Suffix string

const (
	// DataSuffix is a single base directory relative to which user-specific data files should be written.
	// This directory is defined by the environment variable $XDG_DATA_HOME.
	DataSuffix = Suffix("DATA_HOME")

	// ConfigSuffix is a single base directory relative to which user-specific configuration files should be written.
	// This directory is defined by the environment variable $XDG_CONFIG_HOME.
	ConfigSuffix = Suffix("CONFIG_HOME")

	// CacheSuffix is a single base directory relative to which user-specific non-essential (cached) data should be written.
	// This directory is defined by the environment variable $XDG_CACHE_HOME.
	CacheSuffix = Suffix("CACHE_HOME")
)

var userDirBySuffix = map[Suffix]string{
	DataSuffix:   "$HOME/.local/share/%s",
	ConfigSuffix: "$HOME/.config/%s",
	CacheSuffix:  "$HOME/.cache/%s",
}

func defaultDir(vendor string, suffix Suffix) string {
	if dir, ok := userDirBySuffix[suffix]; ok {
		return fmt.Sprintf(dir, vendor)
	}
	return ""
}

func vendorKey(vendor string, suffix Suffix) string {
	return strings.ToUpper(strings.Join([]string{
		vendor, string(suffix),
	}, "_"))
}

// Dir attempts to resolve and respect the XDG Base Directory Specification
// https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
// The resolution to a specific suffix is based on environmental preference
//  1. VENDOR_SUFFIX (ARK_CACHE_HOME)
//  2. XDG_SUFFIX    (XDG_CACHE_HOME)
//  3. DEFAULT       ($HOME/.cache/ark)
func Dir(vendor string, suffix Suffix) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	usrDir := defaultDir(vendor, suffix)
	xdgDir := utils.EnvLookup(vendorKey("XDG", suffix), usrDir)
	vendorDir := utils.EnvLookup(vendorKey(vendor, suffix), xdgDir)
	return fs.NormalizePath(cwd, vendorDir)
}
