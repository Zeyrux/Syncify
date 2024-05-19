package main

import (
	app "syncify/app"
)

// Investigate: LookupError: No results found for song: Luca-Dante Spadafora - Bayrisch Drop
func main() {
	app.CheckVenv()
	app.CheckFFMPEG()
	config := app.NewConfig()
	config.Update()
	config.Save()
}
