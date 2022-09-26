package ark

import (
	"crypto/sha1"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/hex"
	"fmt"
	"hash"
	"os"
	"path/filepath"
	"sort"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/gorm/json_datatypes"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/mitchellh/hashstructure/v2"

	"github.com/myfintech/ark/src/go/lib/fs"
)

type ExcludeFromHash struct {
	SourceFiles json_datatypes.StringSlice `json:"sourceFiles" mapstructure:"sourceFiles"`
}

// Value return json value, implement driver.Valuer interface
func (m ExcludeFromHash) Value() (driver.Value, error) {
	return json_datatypes.MarshalString(&m)
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (m *ExcludeFromHash) Scan(val interface{}) error {
	return json_datatypes.Scan(val, m)
}

// GormDataType gorm common data type
func (m ExcludeFromHash) GormDataType() string {
	return "json"
}

// GormDBDataType gorm db data type
func (ExcludeFromHash) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	return json_datatypes.DetermineDBDataType(db)
}

// RawTarget an object that represents the input of an action
// The RawTarget houses attributes that are easily serializable and must be cast or embedded in to other types
type RawTarget struct {
	ID                       string                            `json:"id" mapstructure:"id" hash:"-"`
	Name                     string                            `json:"name" mapstructure:"name"`
	Type                     string                            `json:"type" mapstructure:"type"`
	File                     string                            `json:"file" mapstructure:"file" hash:"-"`
	Realm                    string                            `json:"realm" mapstructure:"realm" hash:"-"`
	Attributes               json_datatypes.MapStringInterface `json:"attributes" mapstructure:"attributes,remain"`
	SourceFiles              json_datatypes.StringSlice        `json:"sourceFiles" mapstructure:"sourceFiles" hash:"-"`
	Labels                   json_datatypes.StringSlice        `json:"labels" mapstructure:"labels" hash:"-"`
	DependsOn                Ancestors                         `json:"dependsOn" mapstructure:"dependsOn" hash:"-"`
	ExcludeFromHash          ExcludeFromHash                   `json:"excludeFromHash" mapstructure:"excludeFromHash" hash:"-"`
	IgnoreFileNotExistsError bool                              `json:"ignoreFileNotExistsError" mapstructure:"ignoreFileNotExistsError"`
}

// Dir returns the directory of the action file
func (t RawTarget) Dir() string {
	return filepath.Dir(t.File)
}

// Key return the cache key for an action
func (t RawTarget) Key() string {
	return fmt.Sprintf("%s:%s",
		fs.TrimPrefix(t.File, t.Realm), t.Name)
}

// KeyHash return the cache key SHA-1 hash for an action
func (t RawTarget) KeyHash() string {
	rootHash := sha1.New()
	_, _ = fmt.Fprint(rootHash, t.Key())
	return hex.EncodeToString(rootHash.Sum(nil))
}

// RelativeDir return the directory for the target relative to the root
func (t RawTarget) RelativeDir() string {
	return fs.TrimPrefix(t.Realm, t.Dir())
}

// Hashcode implementing the graph hashcode interface
func (t RawTarget) Hashcode() interface{} {
	return t.Key()
}

func (t RawTarget) String() string {
	return t.Key()
}

func (t *RawTarget) validateAndNormalizeSourceFiles() error {
	var normalizedFiles []string
	for idx, file := range t.SourceFiles {
		if file == "" {
			return errors.Errorf("source file at idx[%d] of target %s cannot be empty", idx, t.Key())
		}

		cleanPath, err := fs.NormalizePathByPrefix(file, t.Realm, t.Dir())
		if err != nil {
			return err
		}
		normalizedFiles = append(normalizedFiles, cleanPath)
	}
	t.SourceFiles = normalizedFiles
	return nil
}

// Validate validates specified structs by checking the specified struct fields against their corresponding validation rules
func (t *RawTarget) Validate() error {
	t.ID = t.Key()
	if err := t.validateAndNormalizeSourceFiles(); err != nil {
		return err
	}
	return validation.ValidateStruct(t,
		validation.Field(&t.Name, validation.Required),
		validation.Field(&t.Type, validation.Required),
		validation.Field(&t.File, validation.Required),
		validation.Field(&t.Realm, validation.Required),
	)
}

// Checksum executes the hash calculation algorithm on the provided target
func (t *RawTarget) Checksum() (rootHash hash.Hash, err error) {
	rootHash = sha256.New()

	if err = t.hashAttributes(rootHash); err != nil {
		return
	}

	// TrimPrefixAll returns a sorted list of files
	// add the hash of each file to the root hash
	if err = t.hashSourceFiles(rootHash); err != nil {
		return
	}

	if err = t.hashAncestors(rootHash); err != nil {
		return
	}

	return rootHash, nil
}

func (t *RawTarget) hashAttributes(rootHash hash.Hash) error {
	structHash, err := hashstructure.Hash(t, hashstructure.FormatV2, nil)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(rootHash, "%d\n", structHash)
	if err != nil {
		return err
	}
	return nil
}

func (t *RawTarget) hashAncestors(rootHash hash.Hash) error {
	// sort ancestors
	sort.Slice(t.DependsOn, func(i, j int) bool {
		return t.DependsOn[i].Hash > t.DependsOn[j].Hash
	})

	// add the hash of each ancestor to the rootHash
	for _, ancestor := range t.DependsOn {
		if _, err := fmt.Fprintf(rootHash, "%s\n", ancestor.Hash); err != nil {
			return err
		}
	}
	return nil
}

func (t *RawTarget) hashSourceFiles(rootHash hash.Hash) error {
	for _, file := range fs.TrimPrefixAll(t.SourceFiles, t.Realm) {
		filename := filepath.Join(t.Realm, file)
		stat, err := os.Stat(filename)
		if t.IgnoreFileNotExistsError && os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if stat.IsDir() {
			return errors.Errorf("directories may not be listed as a source_file %s", filename)
		}

		if t.excludeFromSourceFiles(filename) {
			continue
		}

		fileHash, err := fs.HashFile(filename, nil)
		if err != nil {
			return err
		}
		if _, err = fmt.Fprintf(rootHash, "%s:%x\n", file, fileHash.Sum(nil)); err != nil {
			return err
		}
	}
	return nil
}

func (t RawTarget) excludeFromSourceFiles(filename string) bool {
	for _, ignoreFile := range t.ExcludeFromHash.SourceFiles {
		if filename == ignoreFile {
			return true
		}
	}
	return false
}
