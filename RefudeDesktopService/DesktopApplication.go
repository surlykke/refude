package main

import (
	"errors"
	"github.com/surlykke/RefudeServices/common"
	"net/http"
	"fmt"
	"os"
	"syscall"
	"os/exec"
	"strings"
	"regexp"
)

type DesktopApplication struct {
	Type            string
	Version         string `json:",omitempty"`
	Name            string
	GenericName     string `json:",omitempty"`
	NoDisplay       bool
	Comment         string `json:",omitempty"`
	Icon            string `json:",omitempty"`
	Hidden          bool
	OnlyShowIn      common.StringSet
	NotShowIn       common.StringSet
	DbusActivatable bool   `json:",omitempty"`
	TryExec         string `json:",omitempty"`
	Path            string `json:",omitempty"`
	Terminal        bool
	Mimetypes       common.StringSet
	Categories      common.StringSet
	Implements      common.StringSet
	Keywords        common.StringSet
	StartupNotify   bool
	StartupWmClass  string `json:",omitempty"`
	Url             string `json:",omitempty"`
	Actions         map[string]Action
	Id              string
}

type Action struct {
	Comment string
	Name    string
	Exec    string
	Icon    string
}

func (app *DesktopApplication) Data(r *http.Request) (int, string, []byte) {
	if r.Method == "GET" {
		return common.GetJsonData(app)
	} else if r.Method == "POST" {
		action := "_default"
		actionv, ok := r.URL.Query()["action"]
		if ok && len(actionv) > 0{
			action = actionv[0]
		}

		if action,ok := app.Actions[action]; !ok {
			return http.StatusNotAcceptable, "", nil
		} else if argv := strings.Fields(regexp.MustCompile("%[uUfF]").ReplaceAllString(action.Exec, "")); len(argv) == 0 {
			return http.StatusInternalServerError, "", nil
		} else if binary, err := exec.LookPath(argv[0]); err != nil {
			return http.StatusInternalServerError, "", nil
		} else {
			env := os.Environ()
			home := os.Getenv("HOME")
			fmt.Println("home: ", home)
			procAttr := &syscall.ProcAttr{home,  env, []uintptr{}, nil}
			fmt.Println(binary)
			if _, err := syscall.ForkExec(binary, argv, procAttr); err != nil {
				return http.StatusInternalServerError, "", nil
			} else {
				return http.StatusAccepted, "", nil
			}
		}
	} else {
		return http.StatusMethodNotAllowed, "", nil
	}
}

func readDesktopFile(path string) (*DesktopApplication, common.StringSet, error) {
	iniGroups, err := common.ReadIniFile(path)

	if err != nil {
		return nil, nil, err
	}

	if iniGroups[0].Name != "Desktop Entry" {
		return nil, nil, errors.New("Desktop file should start with a 'Desktop Entry' group")
	}

	app := DesktopApplication{}
	desktopEntry := iniGroups[0].Entry
	app.Type = desktopEntry["Type"]
	app.Version = desktopEntry["Version"]
	app.Path = desktopEntry["Path"]
	app.StartupWmClass = desktopEntry["StartupWMClass"]
	app.Url = desktopEntry["URL"]
	app.Icon = desktopEntry["Icon"]

	// FIXME use localized values
	app.Name = desktopEntry["Name"]
	app.GenericName = desktopEntry["GenericName"]
	app.Comment = desktopEntry["Comment"]

	app.OnlyShowIn = common.ToSet(common.Split(desktopEntry["OnlyShowIn"], ";"))
	app.NotShowIn = common.ToSet(common.Split(desktopEntry["NotShowIn"], ";"))
	app.Mimetypes = make(common.StringSet)
	app.Categories = common.ToSet(common.Split(desktopEntry["Categories"], ";"))
	app.Implements = common.ToSet(common.Split(desktopEntry["Implements"], ";"))
	app.Keywords = common.ToSet(common.Split(desktopEntry["Keywords"], ";"))
	app.Actions = make(map[string]Action, 0)
	app.Actions["_default"] = Action{
		Name: app.Name,
		Comment: app.Comment,
		Icon: app.Icon,
		Exec: desktopEntry["Exec"],
	}
	actionNames := common.ToSet(common.Split(desktopEntry["Actions"], ";"))
	for i := 1; i < len(iniGroups); i++ {
		if iniGroups[i].Name[0:15] != "Desktop Action " {
			continue
		} else if actionName := iniGroups[i].Name[15:]; !actionNames[actionName] {
			fmt.Println("Unknown action", iniGroups[i].Name, " - ignoring")
			continue
		} else {
			app.Actions[actionName] = Action{
				Name: iniGroups[i].Entry["Name"],
				Comment: "",
				Icon: iniGroups[i].Entry["Icon"],
				Exec: iniGroups[i].Entry["Exec"],
			}
		}
	}

	return &app, common.ToSet(common.Split(desktopEntry["MimeType"], ";")), nil
}
