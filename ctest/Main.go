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

type Dat struct {
	A string
	B int
}

func main() {
	var dat = Dat{"A", 7}
	var intf interface{} = &dat
	dat.A = "AAA"
	var json1, _ = json.Marshal(dat)
	var json2, _ = json.Marshal(intf)
	fmt.Println("json1:", string(json1))
	fmt.Println("json2:", string(json2))
}
