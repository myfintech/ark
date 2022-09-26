package json_datatypes

import (
	"database/sql/driver"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// StringSlice defined JSON data type, need to implements driver.Valuer, sql.Scanner interface
type StringSlice []string

// Value return json value, implement driver.Valuer interface
func (m StringSlice) Value() (driver.Value, error) {
	return MarshalString(&m)
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (m *StringSlice) Scan(val interface{}) error {
	return Scan(val, m)
}

// GormDataType gorm common data type
func (m StringSlice) GormDataType() string {
	return "json"
}

// GormDBDataType gorm db data type
func (StringSlice) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	return DetermineDBDataType(db)
}
