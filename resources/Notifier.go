package resources

import (
    "net"
    "bufio"
	"fmt"
	"net/http"
	"time"
)


type Notifier struct {
	events     evtChan
	newClients  clientChan
	clients    map[client]bool
}

func NewNotifier() Notifier {
	notifier := Notifier{make(evtChan), make(clientChan), make(map[client]bool)}
	go notifier.run()
	return notifier
}

func (n* Notifier) Notify(eventType string, data string) {
	n.events <- fmt.Sprintf(chunkTemplate, len(eventType) + len(data) + 14, eventType, data)
}


type evtChan chan string
type clientChan chan client
type client struct {
	Conn net.Conn
	W    *bufio.Writer
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

func (n* Notifier) run() {
	for ;; {
		select {
		case client := <-n.newClients:
			if writeAndFlush(client, initialResponse) {
				fmt.Println("Ny klient")
				n.clients[client] = true
			}
		case evt := <-n.events:
			for client,_ := range n.clients {
				if !writeAndFlush(client, evt) {
					fmt.Println("Lukker klient")
					delete(n.clients, client)
				}
			}
		}
	}
}

func writeAndFlush(client client, msg string) bool {
	_, err := client.W.WriteString(msg)
	if err == nil {
		err = client.W.Flush()
	}
	if err != nil {
		client.Conn.Close()
		return false
	} else {
		return true
	}
}



func (n *Notifier) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	hj, ok := w.(http.Hijacker)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	conn, bufrw, err := hj.Hijack()
	if err != nil {
		conn.Close()
		return
	}
	n.newClients <- client{conn, bufrw.Writer}
}

func generateEvents(n *Notifier) {
	for ;; {
		time.Sleep(3 * time.Second)
		n.Notify("tid", string(time.Now().Format(time.UnixDate)))
	}
}

func main() {
	notifier := NewNotifier()
	go generateEvents(&notifier)
	http.ListenAndServe(":8000", &notifier)
}
