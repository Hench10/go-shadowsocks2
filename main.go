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
	"time"

	"./core"
	"net"
)

var config struct {
	Debug      bool
	UDPTimeout time.Duration
}

var logger = log.New(os.Stderr, "", log.Lshortfile|log.LstdFlags)

func logf(f string, v ...interface{}) {
	if config.Debug {
		logger.Output(2, fmt.Sprintf(f, v...))
	}
}

func main() {

	var flags struct {
		Server   string
		Cipher   string
		Password string
		Keygen   int
	}

	flag.BoolVar(&config.Debug, "d", false, "debug mode")
	flag.StringVar(&flags.Cipher, "cipher", "AEAD_CHACHA20_POLY1305", "available ciphers: "+strings.Join(core.ListCipher(), " "))
	flag.IntVar(&flags.Keygen, "keygen", 0, "generate a base64url-encoded random key of given length in byte")
	flag.StringVar(&flags.Password, "password", "", "password")
	flag.StringVar(&flags.Server, "s", "", "server listen address or url")
	flag.DurationVar(&config.UDPTimeout, "udptimeout", 5*time.Minute, "UDP tunnel timeout")
	flag.Parse()

	// 密码生成器
	if flags.Keygen > 0 {
		key := make([]byte, flags.Keygen)
		io.ReadFull(rand.Reader, key) // rand.Reader 密码生成器
		fmt.Println(base64.URLEncoding.EncodeToString(key))
		return
	}

	if flags.Server == "" {
		flag.Usage()
		return
	}

	var key []byte
	addr := flags.Server
	cipher := flags.Cipher
	password := flags.Password
	var err error

	if strings.HasPrefix(addr, "ss://") {
		addr, cipher, password, err = parseURL(addr)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		_,err := net.ResolveTCPAddr("tcp",addr)
		log.Fatal(err)
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
