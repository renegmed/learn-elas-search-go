package main

import "elasticsearch-olivere/web"

func main() {
	r := web.RegisterRoutes()
	r.Static("/public", "./public")
	r.Run(":3000")
}
