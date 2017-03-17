package service

import (
	"fmt"
	"net/http"
	"net"
	"bufio"
)


type evtChan chan []byte

type Notifier struct {
	clients map[evtChan]bool
}

func MakeNotifier() Notifier {
	return Notifier{make(map[evtChan]bool)}
}

func (n Notifier) Notify(eventType string, data string) {
	message := []byte(fmt.Sprintf(chunkTemplate, len(eventType) + len(data) + 14, eventType, data))
	for client,_ := range n.clients {
		select {
		case <-client:
			delete(n.clients, client)
		default:
			client<- message
		}
	}
}

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


func (n *Notifier) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else if hj, ok := w.(http.Hijacker); !ok {
		w.WriteHeader(http.StatusInternalServerError)
	} else if con, bufrw, err := hj.Hijack(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		evts := make(evtChan)
		n.clients[evts] = true
		go HandleClient(evts, con, bufrw)
	}
}

func (n *Notifier) Equal(res Resource) bool {
	return false
}

func HandleClient(evts evtChan, conn net.Conn, bufrw *bufio.ReadWriter) {
	defer func() { evts <- []byte{} }()
	defer conn.Close()

	if _, err := bufrw.Write([]byte(initialResponse)); err == nil {
		for ;; {
			message := <- evts
			if _,err := bufrw.Write([]byte(message)); err != nil || bufrw.Flush() != nil {
				break
			}
		}
	}
}


