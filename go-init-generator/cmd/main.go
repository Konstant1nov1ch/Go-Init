package main

import (
	"context"

	"go-init-gen/internal/app"
)

func main() {
	ctx := context.Background()
	a, err := app.New(ctx)
	if err != nil {
		return
	}
	err = a.Run()
	if err != nil {
		return
	}
}
