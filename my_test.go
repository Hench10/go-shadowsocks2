package main

import (
	"testing"
	"fmt"
	"time"
)

func TestParse(test *testing.T) {
	// s := "aewaea:8888"
	// a, r := net.ResolveTCPAddr("tcp", s)
	// fmt.Println(a, r)
	a := 10

	t1 := time.Duration(a) * time.Minute
	t2 := time.Duration(5 * time.Second)
	fmt.Println(t1, t2)
}
