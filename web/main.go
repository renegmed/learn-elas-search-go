package main

import (
	"embed"
	"log"
)

//go:embed templates
var templates embed.FS

func main() {
	r := RegisterRoutes(&templates)
	r.Static("/public", "./public")
	log.Println("...server started :3000")
	r.Run(":3000")
}
