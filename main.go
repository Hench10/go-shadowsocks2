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
	ManagerAddr   string `json:"manager_addr"`
	ManagerPwd    string `json:"manager_pwd"`
	ManagerMethod string `json:"manager_method"`
}

const (
	TrafficIn  = iota
	TrafficOut
)

type PortInfo struct {
	sync.RWMutex
	Index      int64
	Port       int
	Method     string
	Password   string
	InTraffic  int64
	OutTraffic int64
	TCPConn    net.Listener
	UDPConn    net.PacketConn
}

func (p *PortInfo) AddTraffic(InOut int, t int64) {
	p.Lock()
	defer p.Unlock()

	p.Println()
	switch InOut {
	case TrafficIn:
		p.InTraffic += t
	case TrafficOut:
		p.OutTraffic += t
	}
}

func (p *PortInfo) GetTraffic() (in int64, out int64) {
	p.Lock()
	defer p.Unlock()

	p.Index++
	return p.InTraffic, p.OutTraffic
}

func (p *PortInfo) AddTCP(conn net.Listener) {
	p.Lock()
	defer p.Unlock()
	p.TCPConn = conn
}

func (p *PortInfo) AddUDP(conn net.PacketConn) {
	p.Lock()
	defer p.Unlock()
	p.UDPConn = conn
}

func (p *PortInfo) Println() {
	fmt.Println("Index:",p.Index,"Port:",p.Port,"In:",p.InTraffic,"Out",p.OutTraffic)
}

var config Config
var PortList = make(map[int]*PortInfo)

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
		if strings.HasPrefix(flags.Manager, "ss://") {
			addr, method, password, err = parseURL(flags.Manager)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal("without Prefix ss://")
		}
	} else if flags.Server != "" {
		if strings.HasPrefix(flags.Server, "ss://") {
			addr, method, password, err = parseURL(flags.Server)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal("without Prefix ss://")
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

	addr_tmp, err := net.ResolveTCPAddr("tcp", addr);
	if err != nil {
		log.Fatal(err)
	}
	config.Port = addr_tmp.Port

	if password == "" {
		log.Fatal("password is empty")
	}

	if flags.Manager == "" {
		cipher, err := core.PickCipher(method, key, password)
		if err != nil {
			log.Fatal(err)
		}

		PortList[config.Port] = &PortInfo{
			Index:      0,
			Port:       config.Port,
			Method:     method,
			Password:   password,
			InTraffic:  0,
			OutTraffic: 0,
		}

		go udpRemote(addr, cipher.PacketConn, PortList[config.Port])
		go tcpRemote(addr, cipher.StreamConn, PortList[config.Port])
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
