package main

import (
	"testing"
	"fmt"
	"net"
)

const num = iota

func TestParse(test *testing.T) {
	addr := ":80"
	mm,err := net.ResolveTCPAddr("tcp",addr)
	fmt.Println(mm,err)
}
