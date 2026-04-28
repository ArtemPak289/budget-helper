package main

import (
	"fmt"
	"os"

	"budget-helper/internal/app"
)

func main() {
	err := app.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err.Error())
	}
}
