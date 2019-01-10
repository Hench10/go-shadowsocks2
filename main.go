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
	"sync"
)

type Config struct {
	Debug    bool   `json:"debug"`
	Port     int    `json:"server_port"`
	Password string `json:"password"`
	Method   string `json:"method"`
	Timeout  int    `json:"timeout"`
	// Core     int    `json:"core"`

	// For Reporter
	ManagerMode   bool
	ManagerAddr   string `json:"manager_addr"`
	ManagerPwd    string `json:"manager_pwd"`
	ManagerMethod string `json:"manager_method"`
}

type PortInfo struct {
	sync.RWMutex
	Index    int64
	Port     int
	Method   string
	Password string
	Cipher   core.Cipher
	InTraffic  int64
	OutTraffic  int64
}

func (p *PortInfo) AddTraffic() {
	p.Lock()
	defer p.Unlock()

	println(p.Port)
}

var config Config
var PortList map[int]*PortInfo

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
		if strings.HasPrefix(addr, "ss://") {
			addr, method, password, err = parseURL(addr)
			if err != nil {
				log.Fatal(err)
			}
			config.ManagerMode = true
		}
	} else if flags.Server != "" {
		if strings.HasPrefix(addr, "ss://") {
			addr, method, password, err = parseURL(addr)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		if flags.ConfigFile != "" {
			if err = ParseConfig(flags.ConfigFile, &config); err != nil {
				log.Fatal(err)
			}
		}
		addr = ":" + string(config.Port)
		method = config.Method
		password = config.Password
	}

	if _, err := net.ResolveTCPAddr("tcp", addr); err != nil {
		log.Fatal(err)
	}

	if password == "" {
		log.Fatal("password is empty")
	}

	if flags.Manager == "" {
		cipher, err := core.PickCipher(method, key, password)
		if err != nil {
			log.Fatal(err)
		}

		go udpRemote(addr, cipher.PacketConn)
		go tcpRemote(addr, cipher.StreamConn)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

func parseURL(s string) (addr, method, password string, err error) {
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
