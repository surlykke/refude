// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/surlykke/RefudeServices/lib/xdg"

	"github.com/rakyll/magicmime"
)

type MimeType struct {
	DefaultApp string
}

var client = http.Client{
	Transport: &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", xdg.RuntimeDir+"/org.refude.desktop-service")
		},
	},
}

func getMimetype(mimetypeId string) (*MimeType, error) {
	response, err := client.Get("http://localhost/mimetype/" + mimetypeId)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	var mimetype = &MimeType{}
	var decoder = json.NewDecoder(response.Body)
	err = decoder.Decode(mimetype)
	if err != nil {
		return nil, err
	}

	return mimetype, nil
}

func launchApp(appId, filepath string) error {
	var url = "http://localhost/application/" + appId + "?arg=" + url.QueryEscape(filepath)
	if request, err := http.NewRequest("POST", url, nil); err != nil {
		return err
	} else if response, err := client.Do(request); err != nil {
		return err
	} else {
		defer response.Body.Close()
		return nil
	}
}

func getDefaultApp(mimetypeid string) (string, error) {
	var mimetype, err = getMimetype(mimetypeid)
	if err != nil {
		fmt.Printf("Error getting mimetype : %v", err)
		return "", err
	} else {
		return mimetype.DefaultApp, nil
	}
}

var schemePattern = func() *regexp.Regexp {
	pattern, err := regexp.Compile(`^(\w+):.*$`)
	if err != nil {
		panic(err)
	}
	return pattern
}()

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: RefudeXdgOpen { file | URL}")
	}

	arg := os.Args[1]

	var mimetypeId = ""

	match := schemePattern.FindStringSubmatch(arg)
	if match != nil {
		mimetypeId = "x-scheme-handler/" + match[1]
	} else {
		var err error
		if arg, err = filepath.Abs(arg); err != nil {
			log.Fatal(err)
		}

		if err := magicmime.Open(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK | magicmime.MAGIC_ERROR); err != nil {
			log.Fatal(err)
		}

		defer magicmime.Close()
		mimetypeId, _ = magicmime.TypeByFile(arg)
	}

	log.Println("mimetypeId: ", mimetypeId)

	if len(mimetypeId) == 0 {
		log.Fatal("Could not determine type of " + arg)
	}
	appId, err := getDefaultApp(mimetypeId)
	if err != nil {
		log.Fatal("Error querying default app of ", mimetypeId, err)
	}
	if appId != "" {
		if err = launchApp(appId, arg); err != nil {
			log.Fatal("Error launching " + appId + " with " + arg)
		}
	} else {
		fmt.Println("Calling refudeAppChooser ", arg, mimetypeId)
		exec.Command("refudeAppChooser", arg, mimetypeId).Start()
	}

}
