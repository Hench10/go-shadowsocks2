package webservise

import (
	"github.com/labstack/echo"
)

func midAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		return next(c)
	}
}
