package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"./core"
	"io/ioutil"
	"encoding/json"
	"net"
)

var config Config

type Config struct {
	Debug    bool   `json:"debug"`
	Port     int    `json:"server_port"`
	Password string `json:"password"`
	Method   string `json:"method"`
	Timeout  int    `json:"timeout"`
	// Core     int    `json:"core"`

	// For Reporter
	ManagerAddr   string `json:"manager_addr"`
	ManagerPwd    string `json:"manager_pwd"`
	ManagerMethod string `json:"manager_method"`
}

var logger = log.New(os.Stderr, "", log.Lshortfile|log.LstdFlags)

func logf(f string, v ...interface{}) {
	if config.Debug {
		logger.Output(2, fmt.Sprintf(f, v...))
	}
}

func main() {

	config = Config{
		Debug:   true,
		Port:    1088,
		Method:  "AES-256-CFB",
		Timeout: 600,
	}

	var flags struct {
		ConfigFile string
		Server     string
		Keygen     int
	}

	flag.IntVar(&flags.Keygen, "keygen", 0, "generate a base64url-encoded random key of given length in byte")
	flag.StringVar(&flags.Server, "s", "", "server url like ss://AES-192-CFB:your-password@:8488")
	flag.StringVar(&flags.ConfigFile, "c", "", "config file path")
	flag.BoolVar(&config.Debug, "d", false, "debug mode")
	flag.IntVar(&config.Port, "p", 1066, "server listen port")
	flag.StringVar(&config.Password, "pwd", "", "password")
	flag.StringVar(&config.Method, "m", "AES-192-CFB", "available ciphers: "+strings.Join(core.ListCipher(), " "))
	flag.IntVar(&config.Timeout, "t", 300, "udp timeout")
	flag.Parse()

	// 密码生成器
	if flags.Keygen > 0 {
		key := make([]byte, flags.Keygen)
		io.ReadFull(rand.Reader, key) // rand.Reader 密码生成器
		fmt.Println(base64.URLEncoding.EncodeToString(key))
		return
	}

	fmt.Println(config)
	os.Exit(1)

	var key []byte
	var err error
	var addr, cipher, password string

	if flags.Server != "" {
		if strings.HasPrefix(addr, "ss://") {
			addr, cipher, password, err = parseURL(addr)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		if flags.ConfigFile != ""{
			if err = ParseConfig(flags.ConfigFile,&config);err != nil{
				log.Fatal(err)
			}
		}
		addr = ":" + string(config.Port)
		cipher = config.Method
		password = config.Password
	}

	if _, err := net.ResolveTCPAddr("tcp", addr); err != nil {
		log.Fatal(err)
	}

	if password == "" {
		log.Fatal("password is empty")
	}

	ciph, err := core.PickCipher(cipher, key, password)
	if err != nil {
		log.Fatal(err)
	}

	go udpRemote(addr, ciph.PacketConn)
	go tcpRemote(addr, ciph.StreamConn)



	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

func parseURL(s string) (addr, cipher, password string, err error) {
	u, err := url.Parse(s)
	if err != nil {
		return
	}

	addr = u.Host
	if u.User != nil {
		cipher = u.User.Username()
		password, _ = u.User.Password()
	}
	return
}

func ParseConfig(path string, config interface{}) (err error) {
	file, err := os.Open(path) // For read access.
	if err != nil {
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}

	if err = json.Unmarshal(data, config); err != nil {
		return
	}
	return
}
