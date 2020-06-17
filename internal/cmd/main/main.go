package main

import (
	"fmt"
	"os"

	"github.com/yashap/crius/internal/app"
)

func main() {
	a := app.App{}
	a.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"),
	)

	a.Run(fmt.Sprintf(":%s", os.Getenv("PORT")))
}
