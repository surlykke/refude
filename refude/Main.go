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
	"errors"
	"flag"
	"fmt"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"io/ioutil"
	"net"
	"net/http"
	"net/textproto"
	"os"
	"strings"
)

type HeaderMap map[string]string

func (hm *HeaderMap) String() string {
	var separator = ""
	var res = ""
	for key, val := range *hm {
		res = res + separator + key + ":" + val
		separator = ","
	}
	return res
}

func (hm *HeaderMap) Set(s string) error {
	var tmp = strings.Split(s, ":")
	if len(tmp) != 2 || len(textproto.TrimString(tmp[0])) == 0 || len(textproto.TrimString(tmp[1])) == 0 {
		return errors.New("Header should be of form <key>:<value>")
	} else {
		(*hm)[tmp[0]] = tmp[1]
		return nil
	}
}

func fail(msg string) {
	_,_ = fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func usage() {
	_,_ = fmt.Fprintln(flag.CommandLine.Output(), "Usage: RefudeReq [options] path")
	_,_ = fmt.Fprintln(flag.CommandLine.Output(), "options:")
	flag.PrintDefaults()
	_,_ = fmt.Fprintln(flag.CommandLine.Output(), "path: path to resource (eg. /application/firefox.desktop")
}

func main() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.Usage = usage
	var headerMap = make(HeaderMap)
	flag.Var(&headerMap, "H", "Http header in the form <key>:<value>. May occur multiple times")
	var method = flag.String("X", "GET", "Http method")
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	var client = http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", xdg.RuntimeDir+"/org.refude.desktop-service")
			},
		},
	}
	var url = "http://localhost" + flag.Arg(0)

	var request, err = http.NewRequest(*method, url, nil);
	if err != nil {
		fail(err.Error())
	}
	for key, value := range headerMap {
		request.Header.Set(key, value)
	}

	if response, err := client.Do(request); err != nil {
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
