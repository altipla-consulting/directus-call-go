package callgo

import (
	"encoding/json"
	"fmt"
)

// Partial reads the inner object fields and stores the unknown ones in a map. When marshaled again the unknown
// fields will be kept and any known fields with the same name will override them.
type Partial[T any] struct {
	Value  *T
	Fields map[string]any
}

func (obj *Partial[T]) UnmarshalJSON(data []byte) error {
	obj.Fields = make(map[string]any)
	if err := json.Unmarshal(data, &obj.Fields); err != nil {
		return fmt.Errorf("callgo: cannot decode partial fields: %w", err)
	}
	if err := json.Unmarshal(data, &obj.Value); err != nil {
		return fmt.Errorf("callgo: cannot decode partial value: %w", err)
	}
	return nil
}

func (obj *Partial[T]) MarshalJSON() ([]byte, error) {
	value, err := json.Marshal(obj.Value)
	if err != nil {
		return nil, fmt.Errorf("callgo: cannot encode partial value: %w", err)
	}
	m := make(map[string]any)
	for k, v := range obj.Fields {
		m[k] = v
	}
	if err := json.Unmarshal(value, &m); err != nil {
		return nil, fmt.Errorf("callgo: cannot re-encode partial value: %w", err)
	}
	result, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("callgo: cannot encode partial %w", err)
	}
	return result, nil
}
