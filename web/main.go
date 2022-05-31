package main

import (
	"embed"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//go:embed templates
var templates embed.FS

func main() {
	r := RegisterRoutes(&templates)
	r.Static("/public", "./public")
	http.Handle("/metrics", promhttp.Handler())
	log.Println("...server started :8080")
	r.Run(":8080")
}
