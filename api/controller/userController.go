package controller

import (
	"chat-app/api/handler"
	"github.com/labstack/echo/v4"
)

func SetupUserRoutes(e *echo.Echo) {
	{
		// e.GET("", handler.GetUsers)
		e.POST("/signup", handler.SignUp)
		e.POST("/login", handler.Login)
		e.GET("/users", handler.GetUsers)
		e.GET("/user/:user_id", handler.GetUser)

	}
}
