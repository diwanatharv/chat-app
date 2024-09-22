package main

import (
	"chat-app/api/controller"
	"chat-app/pkg/config"
	middlewares "chat-app/pkg/middleware"
	"os"

	"github.com/labstack/echo/v4/middleware"

	"github.com/labstack/echo/v4"
)

func main() {
	port := os.Getenv("PORT")
	config.InitRedis()
	if port == "" {
		port = "8000"
	}

	e := echo.New()
	e.Use(middleware.Logger())
	go controller.StartRedisSubscription()

	controller.SetupUserRoutes(e)
	e.GET("/ws", controller.HandleWebSocketConnection)
	authRoutes := e.Group("")
	authRoutes.Use(middlewares.AuthenticationMiddleware)

	e.Logger.Fatal(e.Start(":" + port))
}
