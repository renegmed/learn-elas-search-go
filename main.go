package main

import "github.com/renegmed/learn-elas-search-go/web"

func main() {
	r := web.RegisterRoutes()
	r.Static("/public", "./public")
	r.Run(":3000")
}
