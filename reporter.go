package main

import (
	"net"
	"./core"
	"log"
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
}
