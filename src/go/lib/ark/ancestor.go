package ark

import (
	"database/sql/driver"

	"github.com/myfintech/ark/src/go/lib/gorm/json_datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Ancestor represents the key and hash of a parent target
type Ancestor struct {
	Key  string `json:"key" hash:"ignore"`
	Hash string `json:"hash" hash:"ignore"`
}

// Ancestors an slice of ancestors used to define dependencies
type Ancestors []Ancestor

// Value return json value, implement driver.Valuer interface
func (m Ancestors) Value() (driver.Value, error) {
	return json_datatypes.MarshalString(&m)
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (m *Ancestors) Scan(val interface{}) error {
	return json_datatypes.Scan(val, m)
}

// GormDataType gorm common data type
func (m Ancestors) GormDataType() string {
	return "json"
}

// GormDBDataType gorm db data type
func (Ancestors) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	return json_datatypes.DetermineDBDataType(db)
}
