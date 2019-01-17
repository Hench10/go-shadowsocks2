package webservise

import (
	"net"
	"github.com/labstack/echo"
	"net/http"
	"github.com/labstack/echo/middleware"
	"strings"
)

type (
	Manager struct {
		L       net.PacketConn
		Workers map[string]*Worker
		Users   map[string]*User
	}

	Worker struct {
		Addr  net.Addr
		Conn  net.PacketConn
		Ports map[int]*Port
	}

	Port struct {
		P          int
		Method     string
		Password   string
		TrafficIn  int64
		TrafficOut int64
		UserID     string
	}

	User struct {
		UserID     string
		Token      string
		WorkerIP   string
		WorkerPort int
	}
)

var e *echo.Echo
var brain Manager
var workers = make(map[string]*Worker)
var users = make(map[string]*User)
var debug = false

func Start(L net.PacketConn,d bool) {
	debug = d
	brain = Manager{L, workers, users}

	// Echo instance
	e = echo.New()

	// Setting
	e.HideBanner = true
	e.Logger = NewLogger("sys")

	// Trace Log
	if debug {
		e.Logger.SetLevel(DEBUG)
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "[Website]${time_custom} Request:${remote_ip}->${method}:'${uri}' response:${status} ${error}\n",
			CustomTimeFormat: "2006/01/02 15:04:05",
		}))
	}

	// Middleware
	e.Use(middleware.Recover())

	// Route - default
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!\n")
	})

	// Route - admin
	admin := e.Group("/admin", midAuth)
	admin.GET("", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!\n")
	})

	// Socks Server
	go listener()

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

func listener() {
	l := brain.L
	for {
		data := make([]byte, 300)
		_, _, err := l.ReadFrom(data)
		if err != nil {
			e.Logger.Printf("UDP remote listen error: %v", err)
			continue
		}

		command := string(data)
		// var res []byte
		switch {
		case strings.HasPrefix(command, "hello"):
			e.Logger.Info("received hello")
		case strings.HasPrefix(command, "response:"):
			// res = handleRemovePort(bytes.Trim(data[9:], "\x00\r\n "))
			e.Logger.Info("received hello")
		}
	}
}

// func getUser(c echo.Context) error {
// 	// User ID from path `users/:id`
// 	id := c.Param("id")
//  c.QueryParam("team")
//  name := c.FormValue("name")
// 	return c.String(http.StatusOK, id)
// }
