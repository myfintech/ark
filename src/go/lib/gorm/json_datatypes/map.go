package json_datatypes

import (
	"database/sql/driver"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// MapStringInterface defined JSON data type, need to implements driver.Valuer, sql.Scanner interface
type MapStringInterface map[string]interface{}

// Value return json value, implement driver.Valuer interface
func (m MapStringInterface) Value() (driver.Value, error) {
	return MarshalString(&m)
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (m *MapStringInterface) Scan(val interface{}) error {
	return Scan(val, m)
}

// GormDataType gorm common data type
func (m MapStringInterface) GormDataType() string {
	return "json"
}

// GormDBDataType gorm db data type
func (MapStringInterface) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	return DetermineDBDataType(db)
}
