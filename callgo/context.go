package callgo

import (
	"context"
	"encoding/json"
	"fmt"
)

type callGoKey int

const (
	accountabilityKey callGoKey = iota
	rawTriggerKey
)

type Accountability struct {
	User      string `json:"user"`
	Role      string `json:"role"`
	Admin     bool   `json:"admin"`
	App       bool   `json:"app"`
	IP        string `json:"ip"`
	UserAgent string `json:"userAgent"`
	Origin    string `json:"origin"`
}

func AccountabilityFromContext(ctx context.Context) *Accountability {
	return ctx.Value(accountabilityKey).(*Accountability)
}

type TriggerKey struct {
	Value        string
	NumericValue int64
}

func (n *TriggerKey) UnmarshalJSON(data []byte) error {
	// String value, cannot be parsed easily as an integer because it can be an UUID.
	if data[0] == '"' {
		return json.Unmarshal(data, &n.Value)
	}

	// Numeric value that also gets replicated to a string for convenience.
	if err := json.Unmarshal(data, &n.NumericValue); err != nil {
		return err
	}
	n.Value = fmt.Sprintf("%d", n.NumericValue)

	return nil
}

func (n *TriggerKey) String() string {
	return n.Value
}

func (n *TriggerKey) IsEmpty() bool {
	return n.Value == ""
}

type RawTrigger struct {
	Event      string
	Keys       []TriggerKey
	Collection string
	Payload    json.RawMessage

	// Path is the URL path of a manual invokation.
	Path string

	bodyContent []byte
}

func RawTriggerFromContext(ctx context.Context) *RawTrigger {
	return ctx.Value(rawTriggerKey).(*RawTrigger)
}

type Trigger[Payload any] struct {
	Event      string
	Keys       []TriggerKey
	Collection string
	Payload    Payload

	// Path is the URL path of a manual invokation.
	Path string
}

func TriggerFromContext[T any](ctx context.Context) (*Trigger[T], error) {
	raw := RawTriggerFromContext(ctx)
	var payload T
	if err := json.Unmarshal(raw.Payload, &payload); err != nil {
		return nil, fmt.Errorf("callgo: cannot decode trigger payload: %w", err)
	}
	return &Trigger[T]{
		Event:      raw.Event,
		Keys:       raw.Keys,
		Collection: raw.Collection,
		Payload:    payload,
	}, nil
}

func FieldsFromContext[T any](ctx context.Context) (*T, error) {
	var result T
	raw := RawTriggerFromContext(ctx)
	if err := json.Unmarshal(raw.bodyContent, &result); err != nil {
		return nil, fmt.Errorf("callgo: cannot decode trigger fields: %w", err)
	}
	return &result, nil
}
