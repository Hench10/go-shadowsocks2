package main

import (
	"testing"
	"fmt"
	"runtime"
	"compress/zlib"
	"io"
	"bytes"
	"encoding/json"
)

const num = iota

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

var mem runtime.MemStats
var school map[string]*User

func TestParse(test *testing.T) {
	two()
}

func one(){
	runtime.ReadMemStats(&mem)
	a := mem.Alloc

	one := User{"hench", 26}
	two := User{"lucie", 25}
	school = map[string]*User{
		"one": &one,
		"two": &two,
	}

	// for i := 1; i < 1000; i++ {
	// 	school[strconv.Itoa(i)] = &User{"hench", i}
	// }
	runtime.ReadMemStats(&mem)
	b := mem.Alloc

	delete(school, "one")
	delete(school, "two")

	runtime.ReadMemStats(&mem)
	c := mem.Alloc
	fmt.Println("memory", a, b, c)

	l, _ := json.Marshal(school)
	fmt.Println(string(l))
}

func two(){
	var in bytes.Buffer
	b := []byte(`{"8388":{"port":8388,"method":"AES-192-CFB","password":"123456","in":0,"out":0}}{"8388":{"port":8388,"method":"AES-192-CFB","password":"123456","in":0,"out":0}}{"8388":{"port":8388,"method":"AES-192-CFB","password":"123456","in":0,"out":0}}{"8388":{"port":8388,"method":"AES-192-CFB","password":"123456","in":0,"out":0}}{"8388":{"port":8388,"method":"AES-192-CFB","password":"123456","in":0,"out":0}}{"8388":{"port":8388,"method":"AES-192-CFB","password":"123456","in":0,"out":0}}{"8388":{"port":8388,"method":"AES-192-CFB","password":"123456","in":0,"out":0}}{"8388":{"port":8388,"method":"AES-192-CFB","password":"123456","in":0,"out":0}}{"8388":{"port":8388,"method":"AES-192-CFB","password":"123456","in":0,"out":0}}{"8388":{"port":8388,"method":"AES-192-CFB","password":"123456","in":0,"out":0}}`)
	fmt.Println(len(b))
	w := zlib.NewWriter(&in)
	w.Write(b)
	w.Close()
	fmt.Println(len(in.Bytes()))

	var out bytes.Buffer
	r, _ := zlib.NewReader(&in)
	io.Copy(&out, r)
	fmt.Println(len(out.Bytes()))
	fmt.Println(out.String())
}
