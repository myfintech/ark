package json_datatypes

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// MarshalString takes in an interface and marshals it to json and returns it as a Value interface
func MarshalString(m interface{}) (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	jBytes, err := json.Marshal(m)
	return string(jBytes), err
}

// Scan takes in an interface determines it's type and unmarshalls it to a destination interface
func Scan(val interface{}, dest interface{}) error {
	var ba []byte
	switch v := val.(type) {
	case []byte:
		ba = v
	case string:
		ba = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", val))
	}

	err := json.Unmarshal(ba, dest)
	return err
}

// DetermineDBDataType determines the underlying database type implemented by gorm
func DetermineDBDataType(db *gorm.DB) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	}
	return ""
}
