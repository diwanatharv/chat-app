package main

import (
	"chat-app/api/controller"
	"chat-app/api/handler"
	middlewares "chat-app/pkg/middleware"
	"os"

	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"

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

	e.GET("/ws/customer", handler.CustomerWebSocketHandler)
	e.GET("/ws/agent", handler.AgentWebSocketHandler)

	logrus.Infof("Starting server on %s", port)
	e.Logger.Fatal(e.Start(":" + port))
}
