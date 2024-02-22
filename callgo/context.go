package callgo

import (
	"context"
	"encoding/json"
	"fmt"
)

type callgoKey int

const (
	accountabilityKey callgoKey = iota
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
	Value        string `json:"string"`
	NumericValue int64  `json:"number"`
}

func (n *TriggerKey) UnmarshalJSON(data []byte) error {
	if data[0] == '"' {
		return json.Unmarshal(data, &n.Value)
	}
	if err := json.Unmarshal(data, &n.NumericValue); err != nil {
		return err
	}
	n.Value = fmt.Sprintf("%d", n.NumericValue)
	return nil
}

func (n *TriggerKey) String() string {
	return n.Value
}

type RawTrigger struct {
	Event      string     `json:"event"`
	Key        TriggerKey `json:"key"`
	Collection string     `json:"collection"`

	Payload json.RawMessage `json:"payload"`
}

func RawTriggerFromContext(ctx context.Context) *RawTrigger {
	return ctx.Value("trigger").(*RawTrigger)
}

type Trigger[Payload any] struct {
	Event      string     `json:"event"`
	Key        TriggerKey `json:"key"`
	Collection string     `json:"collection"`
	Payload    Payload    `json:"payload"`
}

func TriggerFromContext[T any](ctx context.Context) (*Trigger[T], error) {
	raw := RawTriggerFromContext(ctx)
	var payload T
	if err := json.Unmarshal(raw.Payload, &payload); err != nil {
		return nil, fmt.Errorf("callgo: cannot decode trigger payload: %w", err)
	}
	return &Trigger[T]{
		Event:      raw.Event,
		Key:        raw.Key,
		Collection: raw.Collection,
		Payload:    payload,
	}, nil
}
