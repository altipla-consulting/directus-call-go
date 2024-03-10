package main

import (
	"context"
	"fmt"

	"github.com/altipla-consulting/directus-call-go/callgo"
	"github.com/altipla-consulting/directus-call-go/testapp/subpackage"
)

func main() {
	callgo.Handle(func(ctx context.Context) error {
		fmt.Println("NoParamsNoReturn called")
		return nil
	})

	callgo.Handle(subpackage.NoParamsWithReturnFn)
	callgo.Handle(subpackage.ParamWithReturnFn)
	callgo.Handle(subpackage.AccountabilityFn)
	callgo.Handle(localErrorFn)

	callgo.RegisterMux()
}

func localErrorFn(ctx context.Context) error {
	return fmt.Errorf("error message")
}
