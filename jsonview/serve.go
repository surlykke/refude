package jsonview

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strconv"
)

//go:embed template.html
var	template []byte

var reg = regexp.MustCompile(`"(self|href)"\s*:\s*"http://localhost:7938([^"]*)"`)
var reg2 = regexp.MustCompile(`"(icon)"\s*:\s*"http://localhost:7938([^"]*)"`)

var repl1 = []byte(`"$1": <a href="/jsonview$2">"$2"</a>`)
var repl2 = []byte(`"icon": <a href="$2">"$2"</a>`)

var Handler = &httputil.ReverseProxy{
	Director: func(req *http.Request) {
		fmt.Println("jsonview: req.URL.Host", req.URL.Host, "req.Host:", req.Host )
		req.URL.Scheme = "http"
		req.URL.Host = req.Host
		req.URL.Path = req.URL.Path[len("/jsonview"):]
	},
	ModifyResponse: func(resp *http.Response) error {
		if resp.Header.Get("Content-Type") == "application/vnd.refude+json" {
			var buf bytes.Buffer
			if jsonBytes, err := ioutil.ReadAll(resp.Body); err != nil {
				return err
			} else	if err := json.Indent(&buf, jsonBytes, "", "   "); err != nil {
				return err
			} else {
				var newBody = buf.Bytes()
				newBody = reg.ReplaceAll(newBody, repl1)
				newBody = reg2.ReplaceAll(newBody, repl2)
				newBody = bytes.Replace(template, []byte("@@@@"), newBody, 1)
				resp.Body = io.NopCloser(bytes.NewReader(newBody))
				resp.ContentLength = int64(len(newBody))
				resp.Header.Set("Content-Length", strconv.Itoa(len(newBody)))
				resp.Header.Set("Content-Type", "text/html")
			}
		}
		return nil
	},
}


