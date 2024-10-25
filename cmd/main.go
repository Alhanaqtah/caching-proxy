package main

import (
	"context"
	"log"
	"os"

	"github.com/Alhanaqtah/caching-proxy/internal/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := app.Run(ctx, os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
