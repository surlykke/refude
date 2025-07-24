// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/surlykke/refude/internal/lib/log"
	"github.com/surlykke/refude/internal/lib/utils"
	"github.com/surlykke/refude/internal/lib/xdg"
)

var socketPath = xdg.RuntimeDir + "/org.refude.nm-socket"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-test" {
		runTest()
		os.Exit(0)
	}

	for {
		toStdErr("connecting...")
		if conn, err := net.Dial("unix", socketPath); err != nil {
			time.Sleep(10 * time.Second)
		} else {
			toStdErr("connected...")
			sendCaller(conn)
			go relay(conn, os.Stdin)
			relay(os.Stdout, conn)
			conn.Close()
		}
	}
}

func relay(dst io.Writer, src io.Reader) {
	var buf = make([]byte, 65536)
	for {
		if n, err := src.Read(buf); err != nil {
			return
		} else if _, err := dst.Write(buf[0:n]); err != nil {
			return
		}
	}
}

func sendCaller(conn net.Conn) {
	if parentProcess, err := process.NewProcess(int32(os.Getppid())); err != nil {
		log.Panic("Could not determine parent")
	} else if exe, err := parentProcess.Exe(); err != nil {
		log.Panic("Could not get exe path of parent")
	} else {
		var bufToSend = utils.PrependWithLength([]byte(exe))
		if _, err := conn.Write(bufToSend); err != nil {
			log.Panic("prepend error", err)
		}
	}
}

func toStdErr(v ...any) {
	fmt.Fprintln(os.Stderr, v...)
}

// -------------------------- For Test -----------------------

func runTest() {
	toStdErr("Running as test, launching", os.Args[0], "as subprocess")
	var cmd = exec.Command(os.Args[0])
	cmd.Stderr = os.Stderr
	var cmdStdin, _ = cmd.StdinPipe()
	var cmdStdout, _ = cmd.StdoutPipe()
	go cmd.Run()

	go testRelayStdout(cmdStdout)

	var buf = make([]byte, 65536)
	for {
		n, _ := os.Stdin.Read(buf)
		var tmp = utils.PrependWithLength(buf[0:n])
		cmdStdin.Write(tmp)
	}
}

func testRelayStdout(cmdStdout io.ReadCloser) {
	var buf = make([]byte, 65536)
	for {
		cmdStdout.Read(buf)
		data := utils.StripLength(buf)
		fmt.Println(len(data), ":", string(data))
	}
}
