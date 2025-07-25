// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"os"
	"regexp"
	"strings"

	"github.com/surlykke/refude/internal/lib/utils"
)

const operators = "GET POST PATCH DELETE"

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
	var tmp = utils.Split(s, ":")
	if len(tmp) != 2 || len(textproto.TrimString(tmp[0])) == 0 || len(textproto.TrimString(tmp[1])) == 0 {
		return errors.New("Header should be of form <key>:<value>")
	} else {
		(*hm)[tmp[0]] = tmp[1]
		return nil
	}
}

func fail(msg string) {
	_, _ = fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func usage() {
	_, _ = fmt.Fprintln(flag.CommandLine.Output(), "Usage: RefudeReq [options] path")
	_, _ = fmt.Fprintln(flag.CommandLine.Output(), "options:")
	flag.PrintDefaults()
	_, _ = fmt.Fprintln(flag.CommandLine.Output(), "path: path to resource (eg. /application/firefox.desktop)")
}

/**
 * returns:
 *  - proto and status (as one string)
 *  - response headers
 *  - fully read response body
 *  - error, if any, in which case other return values are nil/zero
 */
func perform(method string, headerMap map[string]string, path string) (string, map[string][]string, []byte, error) {
	var client = &http.Client{}
	var url = "http://localhost:7938" + path

	var request, err = http.NewRequest(method, url, nil)
	if err != nil {
		return "", nil, nil, err
	}

	for key, value := range headerMap {
		request.Header.Set(key, value)
	}

	response, err := client.Do(request)
	if err != nil {
		return "", nil, nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fail(err.Error())
	}

	return response.Proto + " " + response.Status, response.Header, body, nil
}

// Assumes the resource sitting at collectionPath returns a list of strings
func getStringlist(collectionPath string) []string {
	_, _, body, err := perform("GET", nil, collectionPath)
	if err != nil {
		return nil
	}

	var collectionPaths = make([]string, 0, 100)
	err = json.Unmarshal(body, &collectionPaths)
	if err != nil {
		return nil
	}

	return collectionPaths
}

var argReg = regexp.MustCompile(`\s+`)

func completions(argStr string, _ bool) []string {
	var args = argReg.Split(argStr, -1)
	var argc = len(args)
	if argc < 2 {
		return []string{} // Probably never happening
	}

	var lastArg = args[argc-2]
	var curArg = args[argc-1]
	var hasX bool
	for i := 1; i < argc-1; i++ {
		hasX = hasX || "-X" == args[i]
	}

	switch lastArg {
	case "-X":
		return []string{"GET", "POST", "PATCH", "DELETE"}
	case "-H":
		return []string{} // TODO offer common http request headers
	default:
		var comp = make([]string, 0, 2000)
		if strings.HasPrefix("-H", curArg) {
			comp = append(comp, "-H")
		}
		if (!hasX) && strings.HasPrefix("-X", curArg) {
			comp = append(comp, "-X")
		}
		if !strings.HasPrefix(curArg, "-") {
			comp = append(comp, getStringlist("/complete?prefix="+curArg)...)
		}

		return comp
	}
}

func printCompletions(argStr string, filter bool) {
	for _, completion := range completions(argStr, filter) {
		fmt.Println(completion)
	}
}

func main() {
	if len(os.Args) > 2 && ("--_completion" == os.Args[1] || "--_completion_filtered" == os.Args[1]) {
		printCompletions(os.Args[2], os.Args[1] == "--_completion_filtered")
		os.Exit(0)
	}

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

	protoAndStatus, headers, body, err := perform(*method, headerMap, flag.Arg(0))

	if err != nil {
		fail(err.Error())
	}

	fmt.Fprint(os.Stderr, protoAndStatus, "\r\n")
	for name, values := range headers {
		for _, val := range values {
			fmt.Fprint(os.Stderr, name, ":", val, "\r\n")
		}
	}

	fmt.Fprint(os.Stderr, "\r\n")
	fmt.Println(string(body))
}
