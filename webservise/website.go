package webservise

import (
	"net"
	"strings"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"strconv"
	"encoding/json"
	"bytes"
)

type (
	DBConfig struct {
		User     string `json:"user"`
		Password string `json:"password"`
		Host     string `json:"host"`
		Port     string `json:"port"`
		Database string `json:"database"`
	}
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
var db *sql.DB

func Start(L net.PacketConn, dbf DBConfig, d bool) {
	debug = d
	brain = Manager{L, workers, users}

	// Echo instance
	e = echo.New()

	// Setting
	e.Debug = debug
	e.HideBanner = true
	e.Logger = NewLogger("sys")
	e.HTTPErrorHandler = HTTPErrorHandler

	// Part Service
	if err := DBLink(dbf); err != nil {
		e.Logger.Fatal(err)
	}

	// Trace Log
	if e.Debug {
		e.Logger.SetLevel(DEBUG)
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format:           "[Website]${time_custom} Request:${remote_ip}->${method}:'${uri}' response:${status} ${error}\n",
			CustomTimeFormat: "2006/01/02 15:04:05",
		}))
	}

	// Middleware
	e.Use(middleware.Recover())

	// Route - default
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!\n")
	})
	e.GET("/user/:id", getUser)

	// Route - admin
	admin := e.Group("/admin", midAuth)
	admin.GET("", adminIndex)
	admin.GET("/", adminIndex)
	admin.GET("/add/:p", addPort)
	admin.GET("/rm/:p", removePort)
	admin.GET("/ping/:p", pingWorker)
	admin.GET("/pong/:p", pong)

	// Socks Server
	go listener()

	// Start server
	e.Logger.Fatal(e.Start(":81"))
}

func DBLink(dbf DBConfig) (err error) {
	// root:password@tcp(127.0.0.1:3306)/database?charset=utf8
	uri := dbf.User + ":" + dbf.Password + "@tcp(" + dbf.Host + ":" + dbf.Port + ")/" + dbf.Database + "?charset=utf8"
	db, err = sql.Open("mysql", uri)
	if err != nil {
		e.Logger.Fatal(err)
	}
	db.SetMaxOpenConns(2000)
	db.SetMaxIdleConns(1000)
	db.Ping()
	return
}

func listener() {
	l := brain.L
	for {
		data := make([]byte, 300)
		_, addr, err := l.ReadFrom(data)
		if err != nil {
			e.Logger.Printf("UDP remote listen error: %v", err)
			continue
		}

		ip, _ := net.ResolveUDPAddr(addr.String(), addr.Network())
		if n,ok := brain.Workers[string(ip.IP)];!ok{
			brain.Workers[string(ip.IP)] = &Worker{
				Addr:addr,
				Conn  net.PacketConn
				Ports map[int]*Port
			}
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

func getUser(c echo.Context) error {
	var quote string
	id, _ := strconv.Atoi(c.Param("id"))

	row := db.QueryRow("SELECT id, `name` FROM user WHERE id = ?", id)
	err := row.Scan(&id, &quote)

	if err != nil {
		e.Logger.Error(err)
	}

	response := User{UserID: strconv.Itoa(id), Token: quote}
	return c.JSON(http.StatusOK, answer(1, "success", response))
}

func adminIndex(c echo.Context) error {
	return c.File("./static/admin.html")
}

func addPort(c echo.Context) error {
	p := c.Param("p")
	i := strings.Index(p, "@")
	port := p[:i]
	pwd := p[i+1:]

	info := map[string]interface{}{
		"port":     port,
		"password": pwd,
		"method":   "AES-192-CFB",
	}

	var buffer bytes.Buffer
	buffer.Write([]byte("add:"))
	js, _ := json.Marshal(info)
	buffer.Write(js)

	brain.L.WriteTo(buffer.Bytes(), brain.Workers["127.0.0.1"].Addr)
	return c.JSON(http.StatusOK, answer(1, "success", ""))
}

func removePort(c echo.Context) error {
	return c.JSON(http.StatusOK, answer(1, "success", ""))
}

func pingWorker(c echo.Context) error {
	return c.JSON(http.StatusOK, answer(1, "success", ""))
}

func pong(c echo.Context) error {
	return c.JSON(http.StatusOK, answer(1, "success", ""))
}
