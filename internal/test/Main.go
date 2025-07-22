// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/surlykke/refude/internal/lib/log"
)

func get(cmd string, args ...string) []byte {
	if res, err := exec.Command(cmd, args...).Output(); err != nil {
		return []byte{}
	} else {
		return res
	}
}

func collect() []byte {
	var buf = bytes.NewBuffer(make([]byte, 0, 100))
	buf.WriteByte('{')
	buf.Write(get("tmux", "list-clients", "-F", "\"#{client_tty}\":\"#{client_pid}\""))
	buf.Write(get("tmux", "list-sessions", "-F", "\"#{session_id}\":\"#{session_attached_list}\""))
	buf.Write(get("tmux", "list-windows", "-a", "-F", "\"#{window_id}\":\"#{session_id}\""))
	buf.Write(get("tmux", "list-panes", "-a", "-F", "\"#{pane_id}\":\"#{window_id}\""))
	buf.WriteByte('}')
	var res = buf.Bytes()
	var lastPos = len(res) - 2
	for i, b := range res {
		if b == '\n' && i != lastPos {
			res[i] = ','
		}
	}
	return buf.Bytes()
}

func main() {
	var start = time.Now()
	var m = make(map[string]string, 100)
	var bytes = collect()
	err := json.Unmarshal(bytes, &m)
	if err != nil {
		log.Error("unmarshal err:", err)
	}
	var end = time.Now()
	fmt.Println(end.Sub(start))
	for key, val := range m {
		fmt.Println(key, "-->", val)
	}
}
