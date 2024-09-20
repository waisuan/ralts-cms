package main

import (
	"ralts-cms/internal/deps"
	"ralts-cms/internal/httpserver"
	"ralts-cms/internal/httpserver/handlers/machines"
)

func main() {
	d := deps.Initialise()
	h := machines.NewHandler(d)
	svr := httpserver.NewHTTPServer(h)
	err := svr.Start(":1323")
	if err != nil {
		panic(err)
	}
}
