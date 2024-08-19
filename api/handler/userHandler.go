package handler

import (
	"chat-app/api/service"
	"chat-app/pkg/helper"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

func SignUp(c echo.Context) error {
	err := service.SignUp(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": fmt.Sprintf("Failed to sign up: %s", err.Error())})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "User signed up successfully"})
}

func GetUser(c echo.Context) error {
	user, err := service.GetUser(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, user)
}

func Login(c echo.Context) error {
	user, err := service.Login(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, user)
}

func GetUsers(c echo.Context) error {
	if err := helper.CheckUserType(c, "ADMIN"); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	users, err := service.GetUsers(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, users)
}
