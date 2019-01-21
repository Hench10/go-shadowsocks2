package webservise

import (
	"github.com/labstack/echo"
	"net/http"
	"fmt"
)


func midAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		return next(c)
	}
}

func answer(sta int, msg string,data interface{})map[string]interface{}{
	return map[string]interface{}{
		"sta":sta,
		"msg":msg,
		"data":data,
	}
}

func HTTPErrorHandler(err error, c echo.Context) {
	var (
		code = http.StatusInternalServerError
		msg  string
	)

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = "unknow Err Message"
		fmt.Errorf("%v, %v", err, he.Message)
		if he.Internal != nil {
			err = fmt.Errorf("%v, %v", err, he.Internal)
		}
	} else if e.Debug {
		msg = err.Error()
	} else {
		msg = http.StatusText(code)
	}

	c.Logger().Error(msg)

	// Send response
	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead || c.Request().Method == http.MethodGet { // Issue #608
			err = c.File("./static/404.html")
			// err = c.String(code,msg)
		} else {
			err = c.JSON(code, msg)
		}
		if err != nil {
			e.Logger.Error(err)
		}
	}
}