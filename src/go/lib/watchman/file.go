package watchman

import (
	"crypto/sha256"
	"encoding/hex"
)

// File
// https://facebook.github.io/watchman/docs/cmd/query.html
type File struct {
	Name   string `json:"name" mapstructure:"name"`     // Name the filename, relative to the watched root
	Exists bool   `json:"exists" mapstructure:"exists"` // true if the file exists, false if it has been deleted

	CClock string `json:"cclock" mapstructure:"cclock"` // the “created clock”; the clock value when we first observed the file, or the clock value when it last switched from !exists to exists.
	OClock string `json:"oclock" mapstructure:"oclock"` // the “observed clock”; the clock value where we last observed some change in this file or its metadata.

	CTime   int64 `json:"ctime" mapstructure:"ctime"`       // last inode change time measured in integer second
	CTimeMs int64 `json:"ctime_ms" mapstructure:"ctime_ms"` // last inode change time measured in integer millisecond
	CTimeUs int64 `json:"ctime_us" mapstructure:"ctime_us"` // last inode change time measured in integer microsecond
	CTimeNs int64 `json:"ctime_ns" mapstructure:"ctime_ns"` // last inode change time measured in integer nanosecond
	CTimeF  int64 `json:"ctime_f" mapstructure:"ctime_f"`   // last inode change time measured in floating-point second

	MTime   int64 `json:"mtime" mapstructure:"mtime"`       // modified time measured in integer seconds
	MTimeMs int64 `json:"mtime_ms" mapstructure:"mtime_ms"` // modified time measured in integer milliseconds
	MTimeUs int64 `json:"mtime_us" mapstructure:"mtime_us"` // modified time measured in integer microseconds
	MTimeNs int64 `json:"mtime_ns" mapstructure:"mtime_ns"` // modified time measured in integer nanoseconds
	MTimeF  int64 `json:"mtime_f" mapstructure:"mtime_f"`   // modified time measured in floating-point seconds

	Size  int  `json:"size" mapstructure:"size"`   // file size in bytes
	Mode  int  `json:"mode" mapstructure:"mode"`   // file (or directory) mode expressed as a decimal integer
	Uid   int  `json:"uid" mapstructure:"uid"`     // the owning uid
	Gid   int  `json:"gid" mapstructure:"gid"`     // the owning gid
	Ino   int  `json:"ino" mapstructure:"ino"`     // the inode number
	Dev   int  `json:"dev" mapstructure:"dev"`     // the device number
	Nlink int  `json:"nlink" mapstructure:"nlink"` // number of hard links
	New   bool `json:"new" mapstructure:"new"`     // whether this entry is newer than the since generator criteria

	// Since 3.1.
	Type string `json:"type" mapstructure:"type"` // File type. Values listed in the type query expression (https://facebook.github.io/watchman/docs/expr/type.html)

	// Since 4.6.
	SymlinkTarget string `json:"symlink_target" mapstructure:"symlink_target"` // the target of a symbolic link if the file is a symbolic link

	// Since 4.9.
	ContentSHA1Hex string `json:"content.sha1hex" mapstructure:"content.sha1hex"` // the SHA-1 digest of the file’s byte content, encoded as 40 hexadecimal digits (e.g. "da39a3ee5e6b4b0d3255bfef95601890afd80709" for an empty file)
}

// AllFileFields returns a list of all file fields

// BasicFields returns a list of basic fields
func BasicFields() []string {
	return []string{
		"name",
		"exists",
		"new",
		"type",
		"symlink_target",
		"content.sha1hex",
	}
}

// FileHasher file hasher is a complex type for a slice of watchman.File
type FileHasher []File

// CalculateRootHash returns a sha256 of the files
func (fh FileHasher) CalculateRootHash() string {
	rootHash := sha256.New()

	for _, file := range fh {
		rootHash.Write([]byte(file.ContentSHA1Hex))
	}

	return hex.EncodeToString(rootHash.Sum(nil))
}
