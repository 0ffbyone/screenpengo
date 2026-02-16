package main

import (
	"os"
	"screenpengo/internal/app"
)

func main() {
	// Initialize application
	application := app.New()

	// Run the application
	application.Run()

	// Exit gracefully
	os.Exit(0)
}