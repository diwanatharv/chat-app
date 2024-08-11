package middlewares

import (
	"chat-app/pkg/helper"
	"github.com/labstack/echo/v4"
	"net/http"
)

func AuthenticationMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return Authentication(next)
}
func Authentication(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		clientToken := c.Request().Header.Get("token")
		if clientToken == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "No Authorization header provided"})
		}

		claims, err := helper.ValidateToken(clientToken)
		if err != "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": err})
		}

		c.Set("email", claims.Email)
		c.Set("first_name", claims.First_name)
		c.Set("last_name", claims.Last_name)
		c.Set("uid", claims.Uid)

		return next(c)
	}
}
