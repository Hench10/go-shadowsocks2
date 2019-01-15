package main

import (
	"net"
	"fmt"
	"strconv"
	"time"
)

func runManager() {
	addr := ":" + strconv.Itoa(config.ManagerPort)
	fmt.Println("runManager", addr)
	c, err := net.ListenPacket("udp", addr)
	if err != nil {
		logf("UDP remote listen error: %v", err)
		return
	}
	defer c.Close()

	time.Sleep(time.Duration(3)*time.Second)
}
