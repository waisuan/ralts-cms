package main

import (
	"ralts-cms/internal/deps"
	"ralts-cms/internal/httpserver"
	machinesHttp "ralts-cms/internal/httpserver/handlers/machines"
	"ralts-cms/internal/machines"
)

func main() {
	d := deps.Initialise()
	mh := machinesHttp.NewHandler(d)
	mr := machines.NewResolver(d.MachineRepository)
	svr := httpserver.NewHTTPServer(mh, mr)
	err := svr.Start(":1323")
	if err != nil {
		panic(err)
	}
}
