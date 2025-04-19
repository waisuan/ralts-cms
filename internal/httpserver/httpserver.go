package httpserver

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/swaggo/echo-swagger"
	machinesHttp "ralts-cms/internal/httpserver/handlers/machines"
	"ralts-cms/internal/machines"

	_ "github.com/swaggo/echo-swagger/example/docs"
)

func NewHTTPServer(mh *machinesHttp.Handler, mr *machines.Resolver) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Timeout())

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	apiGroup := e.Group("/api")
	machineGroup := apiGroup.Group("/machines")
	machineGroup.GET("/:serialnumber", mh.Get)

	return e
}
