package main

import (
	"log"
	"net"
	"fmt"
	"strconv"
	"./core"
	"./webservise"
)

func runManager() {
	addr := ":" + strconv.Itoa(config.ManagerPort)
	fmt.Println("runManager", addr)

	c, err := net.ListenPacket("udp", addr)
	// defer c.Close()
	if err != nil {
		logf("UDP remote listen error: %v", err)
		return
	}

	var key []byte
	cipher, err := core.PickCipher(config.ManagerMethod, key, config.ManagerPwd)
	if err != nil {
		log.Fatal(err)
	}

	c = cipher.PacketConn(c)

	go webservise.Start(c,config.Debug)
}
