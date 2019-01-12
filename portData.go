package main

import (
	"net"
	"sync"
)

const (
	TrafficIn  = iota
	TrafficOut
)

var PortList = make(map[int]*PortInfo)

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
	logf("Index:", p.Index, "Port:", p.Port, "In:", p.InTraffic, "Out", p.OutTraffic)
}

func NewPort(port int, method, password string) {
	PortList[port] = &PortInfo{
		Index:      0,
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

	PortList[port].Index++
	return PortList[port]
}

func DelPort(port int) {
	if GetPort(port) != nil {
		delete(PortList, port)
	}
}