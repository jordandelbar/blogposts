package main

import (
	"personal_website/cmd/app"
	"personal_website/config"
)

func main() {
	cfg := config.InitConfig()
	app.StartApp(&cfg)
}
