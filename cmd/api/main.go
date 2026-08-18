package main

import (
	"context"
	"log"

	"github.com/NesterovYehor/Inventorio/internal/app"
)

func main() {
	app, err := app.Setup("tmp/tmp.db", ":3000")
	if err != nil {
		log.Println(err)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	if err := app.Run(ctx); err != nil {
		log.Println(err)
	}

}
