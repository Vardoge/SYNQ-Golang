package common

import (
	"encoding/json"
	"errors"
	"fmt"
)

type SynqError struct {
	Name    string           `json:"name"`
	Message string           `json:"message"`
	Url     string           `json:"url"`
	Details *json.RawMessage `json:"details"`
}

func NewError(msg string, args ...interface{}) error {
	m := msg
	if len(args) > 0 {
		m = fmt.Sprintf(msg, args...)
	}
	return errors.New(m)
}
