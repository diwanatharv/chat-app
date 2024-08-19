package main

import (
	"chat-app/api/controller"
	middlewares "chat-app/pkg/middleware"
	"github.com/labstack/echo/v4/middleware"
	"os"

	"github.com/labstack/echo/v4"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	e := echo.New()
	e.Use(middleware.Logger())

	controller.SetupUserRoutes(e)
	authRoutes := e.Group("")
	authRoutes.Use(middlewares.AuthenticationMiddleware)

	e.Logger.Fatal(e.Start(":" + port))
}
