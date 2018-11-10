// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
)

func fail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 4 {
		fail("Not enough arguments")
	}
	var service = os.Args[1]
	var method = os.Args[2]
	var path = os.Args[3]
	var body = bytes.NewBuffer([]byte(strings.Join(os.Args[4:], " ")))

	var client = http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", xdg.RuntimeDir+"/org.refude." + service)
			},
		},
	}
	var url = "http://localhost" + path

	if !slice.Contains([]string{"GET", "POST", "PATCH", "DELETE"}, method) {
		fail("Method " + method + " not supported")
	} else if slice.Contains([]string{"GET", "DELETE"}, method) && body.Len() > 0 {
		panic("No body allowed")
	} else if method == "PATCH" && body.Len() == 0 {
		panic("Body mandatory")
	}

	if request, err := http.NewRequest(method, url, body); err != nil {
		fail(err.Error())
	} else if response, err := client.Do(request); err != nil {
		fail(err.Error())
	} else if body, err := ioutil.ReadAll(response.Body); err != nil {
		fail(err.Error())
	} else {
		fmt.Fprint(os.Stderr, response.Proto, " ", response.Status, "\r\n")
		var isJson bool
		for name, values := range response.Header {
			for _,val := range values {
				fmt.Fprint(os.Stderr, name, ":", val, "\r\n")
				if (name == "Content-Type" && (val == "application/json" || strings.HasSuffix(val, "+json"))) {
					isJson = true
				}
			}
		}
		fmt.Fprint(os.Stderr, "\r\n")
		if isJson {
			var buf bytes.Buffer
			if json.Indent(&buf, body, "", "    "); err == nil {
				body = buf.Bytes()
			}
		}

		fmt.Println(string(body))
	}
}
