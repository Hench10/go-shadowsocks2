package manager

import (
	"github.com/labstack/echo"
)

func midAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := next(c); err != nil {
			c.Error(err)
		}
		return nil
	}
}