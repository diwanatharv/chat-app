package main

import (
	"chat-app/api/controller"
	"chat-app/api/handler"
	middlewares "chat-app/pkg/middleware"
	cache "chat-app/pkg/redis"
	"os"

	"github.com/labstack/echo/v4/middleware"

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
	redisClient := cache.NewRedisClient()
	wsServer := handler.NewWebSocketServer(redisClient)

	go wsServer.StartRedisSubscription()

	authRoutes := e.Group("")
	authRoutes.Use(middlewares.AuthenticationMiddleware)
	// WebSocket route
	e.GET("/ws", func(c echo.Context) error {
		wsServer.HandleConnections(c.Response().Writer, c.Request())
		return nil
	})

	e.Logger.Fatal(e.Start(":" + port))
}
