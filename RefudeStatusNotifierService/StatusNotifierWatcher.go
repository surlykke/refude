package main

import (
	"regexp"
	"errors"
	"github.com/surlykke/RefudeServices/lib/service"
	"github.com/surlykke/RefudeServices/lib/stringlist"
)

var serviceNameReg = regexp.MustCompile(`org.kde.StatusNotifierItem-(.*)`)

func getId(serviceName string) (string, error) {
	m := serviceNameReg.FindStringSubmatch(serviceName)
	if len(m) > 0 {
		return  m[1], nil
	} else {
		return "", errors.New(serviceName + " does not match")
	}
}

func StatusNotifierWatcher(register chan string, unregister chan string) {
	service.Map("/", stringlist.StringList{"ping", "notify", "items/"})
	ids := make(stringlist.StringList, 0)
	service.Map("/items/", ids)

	for {
		select {
		case serviceName := <- register :
			if id, err := getId(serviceName); err == nil {
				ids = stringlist.AppendIfNotThere(ids, id)
				service.Map("/items/", ids)
			}
		case serviceName := <- unregister:
			if id, err := getId(serviceName); err == nil {
				ids = stringlist.Remove(ids, id)
				service.Map("/items/", ids)
			}
		}
	}
}

