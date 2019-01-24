package main

import (
	"net"
	"./core"
	"log"
	"time"
	"bytes"
	"strings"
	"encoding/json"
	"fmt"
	"strconv"
	"compress/zlib"
	"io"
)

type TransMessage struct {
	CMD     string
	Status  bool
	Message string
	Payload string
}

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
	// defer c.Close()

	c = cipher.PacketConn(c)
	m, err := net.ResolveUDPAddr("udp", config.ManagerAddr)

	go reporter(c, m)
	go cmdHandle(c)

	fmt.Println("Connect To Manager ", config.ManagerAddr)
}

func reporter(conn net.PacketConn, manager net.Addr) {
	var buffer bytes.Buffer

	timer := time.Tick(10 * time.Second)
	for {
		select {
		case <-timer:
			data := JsonPort()
			buffer.Reset()
			buffer.Write([]byte("report:"))
			buffer.Write(data)

			if err := send(buffer.Bytes(),conn,manager);err != nil{
				return
			}
			logf(string(buffer.Bytes()))
		}
	}
}

func cmdHandle(conn net.PacketConn) {
	for {
		var tmp bytes.Buffer
		data := make([]byte, 1024)
		n, manager, err := conn.ReadFrom(data)
		if err != nil {
			logf("UDP remote listen error: %v", err)
			return
		}
		tmp.Write(data[:n])

		var rec bytes.Buffer
		r, _ := zlib.NewReader(&tmp)
		io.Copy(&rec, r)

		data2 := rec.Bytes()
		command := string(data)

		var res []byte
		switch {
		case strings.HasPrefix(command, "add:"):
			res = handleAddPort(bytes.Trim(data2[4:], "\x00\r\n "))
		case strings.HasPrefix(command, "remove:"):
			res = handleRemovePort(bytes.Trim(data2[7:], "\x00\r\n "))
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

func handleAddPort(data []byte) (res []byte) {
	var params struct {
		Port     int    `json:"port"`
		Password string `json:"password"`
		Method   string `json:"method"`
	}
	if err := json.Unmarshal(data, &params); err != nil {
		return response("add", false, err.Error(), string(data))
	}
	if params.Port == 0 || params.Password == "" || params.Method == "" {
		return response("add", false, "No Enough Params [port,password,method]", string(data))
	}
	cipher, err := core.PickCipher(params.Method, res, params.Password)
	if err != nil {
		return response("add", false, "New Cipher Failed", string(data))
	}

	NewPort(params.Port, params.Method, params.Password)
	addr := ":" + strconv.Itoa(params.Port)

	go udpRemote(addr, cipher.PacketConn, PortList[params.Port])
	go tcpRemote(addr, cipher.StreamConn, PortList[params.Port])

	return response("add", true, "New Port "+addr+" Success", string(data))
}

func handleRemovePort(data []byte) (res []byte) {
	var params struct {
		Port int `json:"port"`
	}
	if err := json.Unmarshal(data, &params); err != nil {
		return response("remove", false, err.Error(), string(data))
	}

	if params.Port == 0 {
		return response("remove", false, "No Port Provide", string(data))
	}

	DelPort(params.Port)
	return response("remove", true, "Remove Port "+string(params.Port)+" Success", string(data))
}

func response(cmd string, stat bool, msg string, payload string) (res []byte) {
	var info = &TransMessage{
		CMD:     cmd,
		Status:  stat,
		Message: msg,
		Payload: payload,
	}

	var buffer bytes.Buffer
	buffer.Write([]byte("response:"))

	js, _ := json.Marshal(info)
	buffer.Write(js)

	return buffer.Bytes()
}

func send(data []byte,conn net.PacketConn,addr net.Addr)(err error){
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	w.Write(data)
	w.Close()

	_,err = conn.WriteTo(in.Bytes(), addr)
	return
}
