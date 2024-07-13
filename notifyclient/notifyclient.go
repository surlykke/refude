package main

//  #cgo pkg-config: gtk4 gtk4-layer-shell-0
// #include <stdio.h>
// #include <stdlib.h>
// #include "notifyclient.h"
import "C"
import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/r3labs/sse/v2"
	"gopkg.in/cenkalti/backoff.v1"
)

func main() {
	go followFlash()
	C.run()
}

func followFlash() {
	var client *sse.Client
	client = sse.NewClient("http://localhost:7938/watch")
	client.ReconnectStrategy = backoff.NewConstantBackOff(2 * time.Second)

	client.Subscribe("resourceChanged", func(evt *sse.Event) {
		if "resourceChanged" == string(evt.Event) && "/flash" == string(evt.Data) { // Subscribe doesn't filter on Event ??
			getFlash()
		}
	})
}

func getFlash() {
	if resp, err := http.Get("http://localhost:7938/flash"); err != nil {
		closeNotification("Error getting flash", err)
	} else {
		defer resp.Body.Close()
		if body, err := io.ReadAll(resp.Body); err != nil {
			closeNotification("Error reading response", err)
		} else {
			defer resp.Body.Close()
			if len(body) == 0 {
				closeNotification("", nil)
			} else {
				var m = make(map[string]string)
				if err := json.Unmarshal(body, &m); err != nil {
					closeNotification("Error unmarshalling json", err)
				} else {
					var iconFilePath = m["iconFilePath"]
					if "" == iconFilePath {
						C.update(1, C.CString(m["subject"]), C.CString(m["body"]), nil)
					} else {
						C.update(1, C.CString(m["subject"]), C.CString(m["body"]), C.CString(iconFilePath))
					}

				}
			}
		}
	}
}

func closeNotification(msg string, err error) {
	C.update(0, C.CString(""), C.CString(""), nil)
	time.AfterFunc(200*time.Millisecond, func() { C.hide() })
}
