// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
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
