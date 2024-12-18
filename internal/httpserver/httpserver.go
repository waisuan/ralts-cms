package httpserver

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/swaggo/echo-swagger"
	"ralts-cms/graph"
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

	graphqlHandler := handler.NewDefaultServer(
		graph.NewExecutableSchema(
			graph.Config{Resolvers: &graph.Resolver{MachineResolver: mr}},
		),
	)
	e.POST("/query", func(c echo.Context) error {
		graphqlHandler.ServeHTTP(c.Response(), c.Request())
		return nil
	})

	playgroundHandler := playground.Handler("GraphQL", "/query")
	e.GET("/playground", func(c echo.Context) error {
		playgroundHandler.ServeHTTP(c.Response(), c.Request())
		return nil
	})

	return e
}
