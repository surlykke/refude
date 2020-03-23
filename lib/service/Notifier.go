// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package service

// Express that something happened to resource of type resourceType
// at path. What happened may be that it was created, updated or removed.
// A client may discover exactly what by GET'ing the resource


func MakeBrooker

/*import (
	"fmt"
	"net/http"
	"net"
	"github.com/surlykke/RefudeServices/lib/pubsub"
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

var publisher = pubsub.MakePublisher()

// Notify promises that for nextNotification _either_ both message and next are nil _or_ both are non-nil
func Notify(eventType string, data string) {
	message := fmt.Sprintf(chunkTemplate, len(eventType) + len(data) + 14, eventType, data)
	publisher.Publish(message)
}

func ResourceAdded(path string) { Notify("resource-added", path[1:])}
func ResourceUpdated(path string) { Notify("resource-updated", path[1:]) }
func ResourceRemoved(path string) { Notify("resource-removed", path[1:]) }

func client(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Into client, writing initialResponse..")
	if _,err := conn.Write([]byte(initialResponse)); err != nil {
		return
	}

	subscription := publisher.MakeSubscriber()

	for {
		if msg, ok := subscription.Next(); !ok {
			return
		} else if !write(conn, msg.(string)){
			return
		}
	}

}

func write(conn net.Conn, msg string) bool {
	// TODO set some deadline...
	_, err := conn.Write([]byte(msg));
	return err == nil
}

type NotifyResource struct {
}

func (nr* NotifyResource) GET(w http.ResponseWriter, r *http.Request) {
	if hj, ok := w.(http.Hijacker); !ok {
		w.WriteHeader(http.StatusInternalServerError)
	} else if conn, _, err := hj.Hijack(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		go client(conn)
	}
}

*/
