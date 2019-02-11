package main

import (
	"encoding/json"
	"fmt"
)

type Data interface{}

type outer struct {
	A string
	B int
	Inner
}

type Inner struct {
	D string
}

func main() {
	var o = outer{A: "A", B: 7}
	var i = Inner{D: "D"}
	o.Inner = i
	if bytes, err := json.Marshal(o); err == nil {
		fmt.Println(string(bytes))
	} else {
		fmt.Println(err)
	}
}
