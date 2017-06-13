/*
 * Copyright (c) 2017 Christian Surlykke
 *
 * This file is part of the RefudeServices project.
 * It is distributed under the GPL v2 license.
 * Please refer to the GPL2 file for a copy of the license.
 */

package notify

import (
	"fmt"
	"net/http"
	"net"
	"sync"
	"github.com/surlykke/RefudeServices/lib/resource"
)

const initialResponse string =
	"HTTP/1.1 200 OK\r\n" +
	"Connection: keep-alive\r\n" +
	"Content-Type: text/event-stream\r\n" +
	"Transfer-Encoding: chunked\r\n" +
	"\r\n";

const chunkTemplate =
	"%x\r\n"  +   // chunk length in hex
	"event:%s\n"  +
	"data:%s\n" +
	"\n" +
	"\r\n";

type Notification struct {
	message string
	next    *Notification
}

var nextNotification = &Notification{}
var mutex = sync.RWMutex{}
var cond = sync.NewCond(mutex.RLocker())

// Notify promises that for nextNotification _either_ both message and next are nil _or_ both are non-nil
func Notify(eventType string, data string) {
	message := fmt.Sprintf(chunkTemplate, len(eventType) + len(data) + 14, eventType, data)

	mutex.Lock()
	nextNotification.next = &Notification{}
	nextNotification.message = message
	nextNotification = nextNotification.next
	mutex.Unlock()

	cond.Broadcast()
}


func client(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Into client, writing initialResponse..")
	if _,err := conn.Write([]byte(initialResponse)); err != nil {
		return
	}

	mutex.RLock()

	myNextNotification := nextNotification
	for {
		for myNextNotification.message == "" {
			cond.Wait()
		}

		// At this point, both conditions hold:
		//   myNextNotification.message != ""
		//   myNextNotification.nextNotification != nil
		// see Notify
		msg := myNextNotification.message

		mutex.RUnlock()
		if !write(conn, msg) {
			return
		}
		mutex.RLock()

		myNextNotification = myNextNotification.next
	}
}

func write(conn net.Conn, msg string) bool {
	// TODO set some deadline...
	_, err := conn.Write([]byte(msg));
	return err == nil
}


func NotifyGET(this *resource.Resource, w http.ResponseWriter, r *http.Request) {
	if hj, ok := w.(http.Hijacker); !ok {
		w.WriteHeader(http.StatusInternalServerError)
	} else if conn, _, err := hj.Hijack(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		go client(conn)
	}
}

