package ygpt

import (
	"errors"
	"testing"
)

func TestRole_MarshalJSON(t *testing.T) {
	testCases := []struct {
		name     string
		role     Role
		expected string
		err      error
	}{
		{
			name:     "user",
			role:     RoleUser,
			expected: `"User"`,
		},
		{
			name:     "userRu",
			role:     RoleUserRu,
			expected: `"Пользователь"`,
		},
		{
			name:     "assistant",
			role:     RoleAssistant,
			expected: `"Assistant"`,
		},
		{
			name:     "assistantRu",
			role:     RoleAssistantRu,
			expected: `"Ассистент"`,
		},
		{
			name: "unknown",
			role: Role("unknown"),
			err:  ErrMarshalJSON,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.role.MarshalJSON()
			if err != nil {
				if !errors.Is(err, tc.err) {
					t.Fatalf("expected error: %v, got: %v", tc.err, err)
				}
				return
			}

			if tc.err != nil {
				t.Fatalf("expected error, but got nil")
			}

			if s := string(data); s != tc.expected {
				t.Fatalf("expected: %q, got: %q", tc.expected, s)
			}
		})
	}
}

func TestRole_UnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name     string
		data     string
		expected Role
		err      error
	}{
		{
			name:     "user",
			data:     `"User"`,
			expected: RoleUser,
		},
		{
			name:     "userRu",
			data:     `"Пользователь"`,
			expected: RoleUserRu,
		},
		{
			name:     "assistant",
			data:     `"Assistant"`,
			expected: RoleAssistant,
		},
		{
			name:     "assistantRu",
			data:     `"Ассистент"`,
			expected: RoleAssistantRu,
		},
		{
			name: "unknown",
			data: `"unknown"`,
			err:  ErrUnmarshalJSON,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			var role Role
			err := role.UnmarshalJSON([]byte(tc.data))
			if err != nil {
				if !errors.Is(err, tc.err) {
					t.Fatalf("expected error: %v, got: %v", tc.err, err)
				}
				return
			}

			if tc.err != nil {
				t.Fatalf("expected error, but got nil")
			}

			if role != tc.expected {
				t.Fatalf("expected: %v, got: %v", tc.expected, role)
			}
		})
	}
}

func TestModel_MarshalJSON(t *testing.T) {
	testCases := []struct {
		name     string
		model    Model
		expected string
		err      error
	}{
		{
			name:     "general",
			model:    ModelGeneral,
			expected: `"general"`,
		},
		{
			name:  "unknown",
			model: Model("unknown"),
			err:   ErrMarshalJSON,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.model.MarshalJSON()
			if err != nil {
				if !errors.Is(err, tc.err) {
					t.Fatalf("expected error: %v, got: %v", tc.err, err)
				}
				return
			}

			if tc.err != nil {
				t.Fatalf("expected error, but got nil")
			}

			if s := string(data); s != tc.expected {
				t.Fatalf("expected: %q, got: %q", tc.expected, s)
			}
		})
	}
}

func TestModel_UnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name     string
		data     string
		expected Model
		err      error
	}{
		{
			name:     "general",
			data:     `"general"`,
			expected: ModelGeneral,
		},
		{
			name: "unknown",
			data: `"unknown"`,
			err:  ErrUnmarshalJSON,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			var model Model
			err := model.UnmarshalJSON([]byte(tc.data))
			if err != nil {
				if !errors.Is(err, tc.err) {
					t.Fatalf("expected error: %v, got: %v", tc.err, err)
				}
				return
			}

			if tc.err != nil {
				t.Fatalf("expected error, but got nil")
			}

			if model != tc.expected {
				t.Fatalf("expected: %v, got: %v", tc.expected, model)
			}
		})
	}
}
