package main

import (
	"fmt"

	"github.com/mallvielfrass/templater/internal"
)

func main() {
	app := internal.NewApp()
	err := app.Run()
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		panic(err)

	}
}
