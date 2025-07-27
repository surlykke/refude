// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/surlykke/refude/internal/lib/utils"
	"github.com/surlykke/refude/internal/lib/xdg"
)

var callerMessage []byte

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-test" {
		runTest()
		os.Exit(0)
	}

	if parentProcess, err := process.NewProcess(int32(os.Getppid())); err != nil {
		panic("Could not determine parent")
	} else if parentPath, err := parentProcess.Exe(); err != nil {
		panic("Could not get path of parent")
	} else {
		callerMessage = utils.PrependWithLength([]byte(parentPath))
	}
	go relayRefudeToStdout()
	relayStdinToRefude()
}

func relayRefudeToStdout() {
	for {
		if conn, err := net.Dial("unix", xdg.NmSocketPath); err != nil {
			time.Sleep(10 * time.Second)
		} else {
			setRefudeWriter(conn)
			if _, err := conn.Write(callerMessage); err == nil {
				log.Print("Connected to refude")
				io.Copy(os.Stdout, conn)
				log.Print("Disonnected from refude")
			}
			setRefudeWriter(io.Discard)
			conn.Close()
		}
	}
}

func relayStdinToRefude() error {
	var buf = make([]byte, 65536)
	for {
		if n, err := os.Stdin.Read(buf); err == io.EOF {
			os.Exit(0) // Browser has exited
		} else if err != nil {
			log.Print(err)
			os.Exit(1) // Shouldn't happen
		} else {
			getRefudeWriter().Write(buf[0:n])
		}
	}
}

var refudeWriter io.Writer = io.Discard
var refudeWriterLock sync.Mutex

func getRefudeWriter() io.Writer {
	refudeWriterLock.Lock()
	defer refudeWriterLock.Unlock()
	return refudeWriter
}

func setRefudeWriter(w io.Writer) {
	refudeWriterLock.Lock()
	defer refudeWriterLock.Unlock()
	refudeWriter = w
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
		log.Print("runTest, sending", utils.PrependWithLength(buf[0:n]))
		var tmp = utils.PrependWithLength(buf[0:n])
		cmdStdin.Write(tmp)
	}
}

func testRelayStdout(cmdStdout io.ReadCloser) {
	var buf = make([]byte, 65536)
	for {
		cmdStdout.Read(buf)
		data := utils.StripLength(buf)
		fmt.Println("From test: len(data)", ":", string(data))
	}
}
