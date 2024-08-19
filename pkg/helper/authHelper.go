package helper

import (
	"errors"

	"github.com/labstack/echo/v4"
)

func CheckUserType(c echo.Context, role string) error {
	userType := c.Request().Header.Get("user_type")
	if userType != "ADMIN" {
		err := errors.New("unauthorised to access this resource")
		return err
	}
	return nil
}

func MatchUserTypeTOUId(c echo.Context, userId string) error {
	userType := c.Request().Header.Get("user_type")
	uid := c.Get("uid")

	if userType == "USER" && uid != userId {
		err := errors.New("unauthorized to access this resource")
		return err
	}
	err := CheckUserType(c, userType)
	return err
}
