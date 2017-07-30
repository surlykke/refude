// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/surlykke/RefudeServices/lib/ini"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/service"
	"github.com/surlykke/RefudeServices/lib/utils"
	"golang.org/x/text/language"
	"os"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

type DesktopApplication struct {
	Type            string
	Version         string
	Name            ini.LocalizedString
	GenericName     ini.LocalizedString
	NoDisplay       bool
	Comment         ini.LocalizedString
	IconName        string
	IconPath        string
	IconUrl         string
	Hidden          bool
	OnlyShowIn      []string
	NotShowIn       []string
	DbusActivatable bool
	TryExec         string
	Exec            string
	Path            string
	Terminal        bool
	Mimetypes       []string
	Categories      []string
	Implements      []string
	Keywords        ini.LocalizedStringlist
	StartupNotify   bool
	StartupWmClass  string
	Url             string
	Actions         map[string]*Action
	Id              string
	RelevanceHint   int64
	languages       language.Matcher
}

type LocalizedDesktopApplication struct {
	Type            string
	Version         string `json:",omitempty"`
	Name            string
	GenericName     string `json:",omitempty"`
	NoDisplay       bool
	Comment         string `json:",omitempty"`
	IconName        string `json:",omitempty"`
	IconPath        string `json:",omitempty"`
	IconUrl         string `json:",omitempty"`
	Hidden          bool
	OnlyShowIn      []string
	NotShowIn       []string
	DbusActivatable bool   `json:",omitempty"`
	TryExec         string `json:",omitempty"`
	Exec            string `json:",omitempty"`
	Path            string `json:",omitempty"`
	Terminal        bool
	Mimetypes       []string
	Categories      []string
	Implements      []string
	Keywords        []string
	StartupNotify   bool
	StartupWmClass  string `json:",omitempty"`
	Url             string `json:",omitempty"`
	Actions         map[string]*LocalizedAction
	Id              string
	RelevanceHint   int64
}

func (da *DesktopApplication) localize(locale string) *LocalizedDesktopApplication {
	var lda = LocalizedDesktopApplication{
		Type:            da.Type,
		Version:         da.Version,
		Name:            da.Name.LocalOrDefault(locale),
		GenericName:     da.GenericName.LocalOrDefault(locale),
		NoDisplay:       da.NoDisplay,
		Comment:         da.Comment.LocalOrDefault(locale),
		IconName:        da.IconName,
		IconPath:        da.IconPath,
		IconUrl:         da.IconUrl,
		Hidden:          da.Hidden,
		OnlyShowIn:      da.OnlyShowIn,
		NotShowIn:       da.NotShowIn,
		DbusActivatable: da.DbusActivatable,
		TryExec:         da.TryExec,
		Exec:            da.Exec,
		Path:            da.Path,
		Terminal:        da.Terminal,
		Mimetypes:       da.Mimetypes,
		Categories:      da.Categories,
		Implements:      da.Implements,
		Keywords:        da.Keywords.LocalOrDefault(locale),
		StartupNotify:   da.StartupNotify,
		StartupWmClass:  da.StartupWmClass,
		Url:             da.Url,
		Actions:         make(map[string]*LocalizedAction),
		Id:              da.Id,
		RelevanceHint:   da.RelevanceHint,
	}

	for id, action := range da.Actions {
		lda.Actions[id] = action.localize(locale)
	}

	return &lda
}

type Action struct {
	Name     ini.LocalizedString
	Exec     string
	IconName string
	IconPath string
	IconUrl  string
}

type LocalizedAction struct {
	Name     string
	Exec     string
	IconName string
	IconPath string
	IconUrl  string
}

func (a *Action) localize(language string) *LocalizedAction {
	return &LocalizedAction{
		Name:     a.Name.LocalOrDefault(language),
		Exec:     a.Exec,
		IconName: a.IconName,
		IconPath: a.IconPath,
		IconUrl:  a.IconUrl,
	}
}

type IconPath string

func (ip IconPath) GET(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, string(ip))
}

func makeDesktopApplication() DesktopApplication {
	var da = DesktopApplication{
		Name:        make(ini.LocalizedString),
		GenericName: make(ini.LocalizedString),
		Comment:     make(ini.LocalizedString),
		OnlyShowIn:  []string{},
		NotShowIn:   []string{},
		Mimetypes:   []string{},
		Categories:  []string{},
		Implements:  []string{},
		Keywords:    make(ini.LocalizedStringlist),
		Actions:     make(map[string]*Action),
		languages:   nil,
	}
	return da
}

func (da *DesktopApplication) Copy() *DesktopApplication {
	cp := *da
	cp.Name = cp.Name.Copy()
	cp.GenericName = cp.GenericName.Copy()
	cp.Comment = cp.Comment.Copy()
	cp.OnlyShowIn = utils.Copy(cp.OnlyShowIn)
	cp.NotShowIn = utils.Copy(cp.NotShowIn)
	cp.Mimetypes = utils.Copy(cp.Mimetypes)
	cp.Categories = utils.Copy(cp.Categories)
	cp.Implements = utils.Copy(cp.Implements)
	cp.Keywords = cp.Keywords.Copy()
	cp.Actions = make(map[string]*Action)
	for id, action := range da.Actions {
		cp.Actions[id] = action.Copy()
	}

	return &cp
}

func makeAction() Action {
	return Action{Name: make(ini.LocalizedString)}
}

func (a *Action) Copy() *Action {
	copy := *a
	copy.Name = copy.Name.Copy()
	return &copy
}

func (da *DesktopApplication) GET(w http.ResponseWriter, r *http.Request) {
	locale := getPreferredLocale(r, da.languages)
	resource.JsonGET(da.localize(locale), w)
}

func getPreferredLocale(r *http.Request, matcher language.Matcher) string {
	if acceptLanguage := r.Header.Get("Accept-Language"); acceptLanguage != "" {
		if tags, _, err := language.ParseAcceptLanguage(acceptLanguage); err == nil {
			tag, _, confidence := matcher.Match(tags...)
			if confidence > language.Low {
				return tag.String()
			}
		}
	}
	return ""
}

func (da *DesktopApplication) POST(w http.ResponseWriter, r *http.Request) {

	actionId := resource.GetSingleQueryParameter(r, "action", "")
	fmt.Print("actionId = '", actionId, "'\n")
	var exec string
	if actionId != "" {
		if action, ok := da.Actions[actionId]; !ok {
			w.WriteHeader(http.StatusNotAcceptable)
			return
		} else {
			exec = action.Exec
		}
	} else {
		exec = da.Exec
	}
	fmt.Println("exec: ", exec)
	var args = strings.Join(r.URL.Query()["arg"], " ")
	var argvAsString = regexp.MustCompile("%[uUfF]").ReplaceAllString(exec, args)
	fmt.Println("Running cmd: " + argvAsString)
	if err := runCmd(da.Terminal, strings.Fields(argvAsString)); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusAccepted)
		updatedApp := da.Copy()
		updatedApp.RelevanceHint = time.Now().UnixNano() / 1000000
		service.Map(r.URL.Path, updatedApp)
	}

}

func runCmd(runInTerminal bool, argv []string) error {
	var cmd *exec.Cmd
	if runInTerminal {
		var terminal, ok = os.LookupEnv("TERMINAL")
		if !ok {
			return errors.New("Trying to run " + strings.Join(argv, " ") + " in terminal, but env variable TERMINAL not set")
		}
		argv = append([]string{terminal, "-e"}, argv...)
	}
	cmd = exec.Command(argv[0], argv[1:]...)

	cmd.Dir = xdg.Home
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return err
	}

	/*if err := cmd.Wait(); err != nil {
		return err
	}*/

	return nil
}


func readDesktopFile(path string) (*DesktopApplication, error) {
	fmt.Println("Reading desktopFile: ", path)
	if iniFile, err := ini.ReadIniFile(path); err != nil {
		return nil, err
	} else if len(iniFile) == 0 || iniFile[0].Name != "Desktop Entry" {
		return nil, errors.New("File must start with '[Desktop Entry]'")
	} else {
		var da = makeDesktopApplication()
		var actionNames = []string{}

		group := iniFile[0]
		if _type, ok := group.Entries["Type"]; ok {
			da.Type = _type[""]
		} else {
			return nil, errors.New("Desktop file invalid, no 'Type' given")
		}
		da.Version = group.Entries["Version"][""]
		if name, ok := group.Entries["Name"]; ok {
			da.Name = name
		} else {
			return nil, errors.New("Desktop file invalid, no 'Name' given")
		}
		da.Name = group.Entries["Name"]
		da.GenericName = group.Entries["GenericName"]
		da.NoDisplay = group.Entries["NoDisplay"][""] == "true"
		da.Comment = group.Entries["Comment"]
		icon := group.Entries["Icon"][""]
		if strings.HasPrefix(icon, "/") {
			da.IconPath = icon
			da.IconUrl = "../icons" + icon
		} else {
			da.IconName = icon
		}
		da.Hidden = group.Entries["Hidden"][""] == "true"
		da.OnlyShowIn = utils.Split(group.Entries["OnlyShowIn"][""], ";")
		da.NotShowIn = utils.Split(group.Entries["NotShowIn"][""], ";")
		da.DbusActivatable = group.Entries["DBusActivatable"][""] == "true"
		da.TryExec = group.Entries["TryExec"][""]
		da.Exec = group.Entries["Exec"][""]
		da.Path = group.Entries["Path"][""]
		da.Terminal = group.Entries["Terminal"][""] == "true"
		actionNames = utils.Split(group.Entries["Actions"][""], ";")
		da.Mimetypes = utils.Split(group.Entries["MimeType"][""], ";")
		da.Categories = utils.Split(group.Entries["Categories"][""], ";")
		da.Implements = utils.Split(group.Entries["Implements"][""], ";")
		// FIXMEda.Keywords[tag] = utils.Split(group[""], ";")
		da.StartupNotify = group.Entries["StartupNotify"][""] == "true"
		da.StartupWmClass = group.Entries["StartupWMClass"][""]
		da.Url = group.Entries["URL"][""]

		for _, actionGroup := range iniFile[1:] {
			if !strings.HasPrefix(actionGroup.Name, "Desktop Action ") {
				log.Print(path, ", ", "Unknown group type: ", actionGroup.Name, " - ignoring\n")
			} else if currentAction := actionGroup.Name[15:]; !utils.Contains(actionNames, currentAction) {
				log.Print(path, ", undeclared action: ", currentAction, " - ignoring\n")
			} else {
				fmt.Println("ActionGroup: ", actionGroup)
				var action = makeAction()
				if name,ok := actionGroup.Entries["Name"]; ok {
					action.Name = name
				} else {
					return nil, errors.New("Desktop file invalid, action " + actionGroup.Name + " has no default 'Name'")
				}
				icon = actionGroup.Entries["Icon"][""]
				if strings.HasPrefix(icon, "/") {
					action.IconPath = icon
					action.IconUrl = "../icons" + icon
				} else {
					action.IconName = icon
				}
				action.Exec = actionGroup.Entries["Exec"][""]
				da.Actions[currentAction] = &action
			}
		}


		for _, action := range da.Actions {
			if action.IconName == "" && action.IconPath == "" {
				action.IconName = da.IconName
				action.IconPath = da.IconPath
				action.IconUrl = da.IconUrl
			}
		}

		var collectedLanguages = make(map[string]bool)
		for locale, _ := range da.Name {
			collectedLanguages[locale] = true
		}
		for locale, _ := range da.GenericName {
			collectedLanguages[locale] = true
		}
		for locale, _ := range da.Comment {
			collectedLanguages[locale] = true
		}
		for locale, _ := range da.Keywords {
			collectedLanguages[locale] = true
		}

		for _, action := range da.Actions {
			for locale, _ := range action.Name {
				collectedLanguages[locale] = true
			}
		}

		var tags = make([]language.Tag, len(collectedLanguages))
		for locale, _ := range collectedLanguages {
			tags = append(tags, language.Make(locale))
		}
		da.languages = language.NewMatcher(tags)

		return &da, nil
	}
}

func transformLanguageTag(tag string) string {
	return strings.Replace(strings.Replace(tag, "_", "-", -1), "@", "-", -1)
}
