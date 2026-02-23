package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONMap represents a JSONB object column.
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(src interface{}) error {
	if src == nil {
		*j = make(JSONMap)
		return nil
	}
	source, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONMap", src)
	}
	return json.Unmarshal(source, j)
}

// JSONArray represents a JSONB array column.
type JSONArray []interface{}

func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	return json.Marshal(j)
}

func (j *JSONArray) Scan(src interface{}) error {
	if src == nil {
		*j = make(JSONArray, 0)
		return nil
	}
	source, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONArray", src)
	}
	return json.Unmarshal(source, j)
}
