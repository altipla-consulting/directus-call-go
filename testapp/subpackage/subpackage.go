package subpackage

import (
	"context"
	"fmt"

	"github.com/altipla-consulting/directus-call-go/callgo"
)

func NoParamsWithReturnFn(ctx context.Context) (string, error) {
	return "foo-value", nil
}

type fooExample struct {
	Foo string `json:"foo"`
	Bar int32  `json:"bar"`
}

func ParamWithReturnFn(ctx context.Context, foo *fooExample) (*fooExample, error) {
	foo.Foo += "new-foo-value"
	foo.Bar = 42
	return foo, nil
}

func AccountabilityFn(ctx context.Context) error {
	fmt.Printf("%#v\n", callgo.AccountabilityFromContext(ctx))
	return nil
}

func FailedValidationErrorFn(ctx context.Context) error {
	return callgo.NewFailedValidationError("foo", "foo", "This field is not accepted")
}

func InvalidErrorFn(ctx context.Context) error {
	return callgo.NewInvalidError("This is an invalid error")
}
