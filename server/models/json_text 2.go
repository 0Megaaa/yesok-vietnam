package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type JSONText json.RawMessage

func (j JSONText) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	if !json.Valid(j) {
		return nil, errors.New("invalid json")
	}
	return string(j), nil
}

func (j *JSONText) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return nil
	}

	if len(b) == 0 {
		*j = nil
		return nil
	}

	if !json.Valid(b) {
		*j = nil
		return nil
	}

	*j = append((*j)[0:0], b...)
	return nil
}

func (j JSONText) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

func (j *JSONText) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*j = nil
		return nil
	}
	if !json.Valid(b) {
		return errors.New("invalid json")
	}
	*j = append((*j)[0:0], b...)
	return nil
}
