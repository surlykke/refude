package main

import (
	"net/http"
	"github.com/surlykke/RefudeServices/common"
)

type Rect struct {
	X,Y int16
	W,H uint16
}

type Display struct {
	W,H uint16
	Screens []Rect
}

func (d* Display) Data(r *http.Request) (int, string, []byte) {
	if r.Method == "GET" {
		return common.GetJsonData(d)
	} else {
		return http.StatusMethodNotAllowed, "", nil
	}
}
