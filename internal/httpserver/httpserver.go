package httpserver

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"ralts-cms/internal/httpserver/handlers/machines"
)

func NewHTTPServer(h *machines.Handler) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Timeout())

	apiGroup := e.Group("/api")
	machineGroup := apiGroup.Group("/machines")
	machineGroup.GET("/:serialnumber", h.Get)

	return e
}
