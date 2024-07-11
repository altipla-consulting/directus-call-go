package callgo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type dynamicTest struct {
	Foo string `json:"foo"`
	Bar int32  `json:"bar"`
}

func TestPartialUnmarshal(t *testing.T) {
	var obj Partial[dynamicTest]
	require.NoError(t, obj.UnmarshalJSON([]byte(`{"foo":"foo-value","bar":42,"baz":"baz-value","deep":{"object":true}}`)))

	require.Equal(t, obj.Value.Foo, "foo-value")
	require.EqualValues(t, obj.Value.Bar, 42)

	require.Equal(t, obj.Fields, map[string]any{
		"foo": "foo-value",
		"bar": float64(42),
		"baz": "baz-value",
		"deep": map[string]any{
			"object": true,
		},
	})
}

func TestPartialMarshal(t *testing.T) {
	obj := Partial[dynamicTest]{
		Value: &dynamicTest{
			Foo: "foo-value",
			Bar: 42,
		},
		Fields: map[string]any{
			"foo": "foo-value-old",
			"bar": float64(45),
			"baz": "baz-value",
			"deep": map[string]any{
				"object": true,
			},
		},
	}
	data, err := obj.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, `{"foo":"foo-value","bar":42,"baz":"baz-value","deep":{"object":true}}`, string(data))
}
