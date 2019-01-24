package main

import (
	"net"
	"sync"
	"encoding/json"
)

const (
	TrafficIn  = iota
	TrafficOut
)

var opIndex int
var PortList = make(map[int]*PortInfo)

type PortInfo struct {
	sync.RWMutex
	Port       int            `json:"port"`
	Method     string         `json:"method"`
	Password   string         `json:"password"`
	InTraffic  int64          `json:"in"`
	OutTraffic int64          `json:"out"`
	TCPConn    net.Listener   `json:"-"`
	UDPConn    net.PacketConn `json:"-"`
}

func (p *PortInfo) AddTraffic(InOut int, t int64) {
	p.Lock()
	defer p.Unlock()

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
	logf("Port:", p.Port, "In:", p.InTraffic, "Out", p.OutTraffic)
}

func NewPort(port int, method, password string) {
	DelPort(port)
	PortList[port] = &PortInfo{
		Port:       port,
		Method:     method,
		Password:   password,
		InTraffic:  0,
		OutTraffic: 0,
	}
}

func GetPort(port int) *PortInfo {
	if _, ok := PortList[port]; !ok {
		return nil
	}
	return PortList[port]
}

func DelPort(port int) {
	if p := GetPort(port); p != nil {
		p.UDPConn.Close()
		p.TCPConn.Close()
		delete(PortList, port)
	}
}

func JsonPort() []byte {
	opIndex++
	var tmp = map[string]interface{}{
		"index":opIndex,
		"ports":PortList,
	}
	data, err := json.Marshal(tmp)
	if err != nil || len(data) == 0 {
		return []byte("")
	}
	return data
}
