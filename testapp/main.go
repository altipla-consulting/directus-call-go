package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/altipla-consulting/directus-call-go/callgo"
)

type fooExample struct {
	Foo string `json:"foo"`
	Bar int32  `json:"bar"`
}

func main() {
	callgo.Handle("NoParamsNoReturn", func(ctx context.Context) error {
		fmt.Println("NoParamsNoReturn called")
		return nil
	})

	callgo.Handle("NoParamsWithReturn", func(ctx context.Context) (string, error) {
		return "foo-value", nil
	})

	callgo.Handle("ParamWithReturn", func(ctx context.Context, foo *fooExample) (*fooExample, error) {
		foo.Foo += "new-foo-value"
		foo.Bar = 42
		return foo, nil
	})

	callgo.Handle("Accountability", func(ctx context.Context) error {
		fmt.Printf("%#v\n", callgo.AccountabilityFromContext(ctx))
		return nil
	})

	callgo.Handle("Error", func(ctx context.Context) error {
		return fmt.Errorf("error message")
	})

	callgo.Serve(callgo.WithLogger(slog.Default()))
}
