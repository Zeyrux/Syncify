package main

import (
	app "syncify/app"
)

func main() {
	app.CheckVenv()
	app.CheckFFMPEG()
	config := app.NewConfig()
	config.Update()
	config.Save()
}
