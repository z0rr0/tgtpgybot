package ygpt

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	// ErrMarshalJSON is an error that occurs when a custom type is marshaled to JSON.
	ErrMarshalJSON = errors.New("failed JSON marshal")

	// ErrUnmarshalJSON is an error that occurs when a custom type is unmarshaled from JSON.
	ErrUnmarshalJSON = errors.New("failed JSON unmarshal")
)

// Model is a type of LLM model.
type Model string

// ModelGeneral is a general LLM model.
const ModelGeneral Model = "general"

// MarshalJSON implements the json.Marshaler interface.
func (m *Model) MarshalJSON() ([]byte, error) {
	return marshalJSON(m, ModelGeneral)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (m *Model) UnmarshalJSON(b []byte) error {
	return unMarshalJSON(m, b, ModelGeneral)
}

// Role is a type of user message role.
type Role string

// User message roles.
const (
	RoleUser        Role = "User"
	RoleUserRu      Role = "Пользователь"
	RoleAssistant   Role = "Assistant"
	RoleAssistantRu Role = "Ассистент"
)

// MarshalJSON implements the json.Marshaler interface.
func (r *Role) MarshalJSON() ([]byte, error) {
	return marshalJSON(r, RoleUser, RoleAssistant, RoleUserRu, RoleAssistantRu)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (r *Role) UnmarshalJSON(b []byte) error {
	return unMarshalJSON(r, b, RoleUser, RoleAssistant, RoleUserRu, RoleAssistantRu)
}

// StringCommonType is a generic interface for custom string based types.
type StringCommonType interface {
	Model | Role
}

// marshalJSON is a generic function for custom types JSON marshal.
func marshalJSON[T StringCommonType](t *T, values ...T) ([]byte, error) {
	var (
		v = *t
		s = string(v)
	)

	for _, value := range values {
		if v == value {
			return json.Marshal(s)
		}
	}

	return nil, errors.Join(ErrMarshalJSON, fmt.Errorf("invalid value: %v", v))
}

// unMarshalJSON is a generic function for custom types JSON unmarshal.
func unMarshalJSON[T StringCommonType](t *T, b []byte, values ...T) error {
	var s string

	if err := json.Unmarshal(b, &s); err != nil {
		return errors.Join(ErrUnmarshalJSON, fmt.Errorf("invalid string value: %v", string(b)))
	}

	v := T(s)

	for _, value := range values {
		if v == value {
			*t = v
			return nil
		}
	}

	return errors.Join(ErrUnmarshalJSON, fmt.Errorf("invalid value: %v", v))
}
