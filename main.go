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
	"errors"
)

type Config struct {
	Debug    bool   `json:"debug"`
	Port     int    `json:"server_port"`
	Password string `json:"password"`
	Method   string `json:"method"`
	Timeout  int    `json:"timeout"`
	// Core     int    `json:"core"`

	// For Manage Mode
	ManagerAddr   string `json:"manager_addr"` //only Reporter
	ManagerPort   string `json:"manager_port"` // only Manager
	ManagerPwd    string `json:"manager_pwd"`
	ManagerMethod string `json:"manager_method"`
}

var config Config

var logger = log.New(os.Stderr, "", log.Lshortfile|log.LstdFlags)

func logf(f string, v ...interface{}) {
	if config.Debug {
		logger.Output(2, fmt.Sprintf(f, v...))
	}
}

func main() {
	var flags struct {
		ConfigFile string
		Server     string
		Manager    string
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
	flag.StringVar(&flags.Manager, "ss", "", "server url like ss://AES-192-CFB:your-password@192.168.1.10:8488")
	flag.Parse()

	// 密码生成器
	if flags.Keygen > 0 {
		key := make([]byte, flags.Keygen)
		io.ReadFull(rand.Reader, key) // rand.Reader 密码生成器
		fmt.Println(base64.URLEncoding.EncodeToString(key))
		return
	}

	var key []byte
	var err error
	var addr, method, password string

	if flags.Manager != "" {
		if addr, method, password, err = parseURL(flags.Manager); err != nil {
			log.Fatal(err)
		}
	} else if flags.Server != "" {
		if addr, method, password, err = parseURL(flags.Server); err != nil {
			log.Fatal(err)
		}
	} else {
		if flags.ConfigFile != "" {
			if err = parseConfig(flags.ConfigFile, &config); err != nil {
				log.Fatal(err)
			}
		}
		addr = ":" + string(config.Port)
		method = config.Method
		password = config.Password
	}

	addrTmp, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	config.Port = addrTmp.Port

	if password == "" {
		log.Fatal("password is empty")
	}

	if flags.Manager == "" {
		cipher, err := core.PickCipher(method, key, password)
		if err != nil {
			log.Fatal(err)
		}

		NewPort(config.Port, method, password)

		go udpRemote(addr, cipher.PacketConn, PortList[config.Port])
		go tcpRemote(addr, cipher.StreamConn, PortList[config.Port])
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

func parseURL(s string) (addr, method, password string, err error) {
	if !strings.HasPrefix(s, "ss://") {
		return addr, method, password, errors.New("Addr without 'ss://' prefix ")
	}

	u, err := url.Parse(s)
	if err != nil {
		return
	}

	addr = u.Host
	if u.User != nil {
		method = u.User.Username()
		password, _ = u.User.Password()
	}
	return
}

func parseConfig(path string, config interface{}) (err error) {
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


