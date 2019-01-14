package main

import (
	"net"
	"./core"
	"log"
	"time"
	"bytes"
	"strings"
)

func runReporter() {
	var key []byte
	cipher, err := core.PickCipher(config.ManagerMethod, key, config.ManagerPwd)
	if err != nil {
		log.Fatal(err)
	}

	c, err := net.ListenPacket("udp", "")
	if err != nil {
		logf("UDP remote listen error: %v", err)
		return
	}
	defer c.Close()

	c = cipher.PacketConn(c)
	m, err := net.ResolveUDPAddr("udp", config.ManagerAddr)

	go reporter(c, m)
	go cmdHandle(c)
}

func reporter(conn net.PacketConn, manager net.Addr) {
	var empty = []byte("ping")
	var buffer bytes.Buffer
	buffer.Write([]byte("report:"))

	timer := time.Tick(10 * time.Second)
	for {
		select {
		case <-timer:
			data := JsonPort()
			if len(data) == 0 {
				conn.WriteTo(empty, manager)
			}
			buffer.Write(data)
			conn.WriteTo(buffer.Bytes(), manager)
		}
	}
}

func cmdHandle(conn net.PacketConn) {
	for {
		data := make([]byte, 300)
		_, manager, err := conn.ReadFrom(data)

		if err != nil {
			logf("UDP remote listen error: %v", err)
			continue
		}

		command := string(data)
		var res []byte
		switch {
		case strings.HasPrefix(command, "add:"):
			res = handleAddPort(bytes.Trim(data[4:], "\x00\r\n "))
		case strings.HasPrefix(command, "remove:"):
			res = handleRemovePort(bytes.Trim(data[7:], "\x00\r\n "))
		case strings.HasPrefix(command, "ping"):
			res = []byte("pong")
		}
		if len(res) == 0 {
			continue
		}
		_, err = conn.WriteTo(res, manager)
		if err != nil {
			logf("Failed to write UDP manage msg, error: ", err.Error())
			continue
		}
	}
}
