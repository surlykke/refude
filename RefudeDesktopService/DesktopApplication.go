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
	"io/ioutil"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"
	"bufio"
	"os"
	"log"
	"golang.org/x/text/language"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/service"
	"github.com/surlykke/RefudeServices/lib/utils"
)

type localizedString map[string]string       // Map from a locale - eg. da_DK - to a string

func (ls localizedString) copy() localizedString {
	var res localizedString = make(localizedString)
	for locale, str := range ls {
		res[locale] = str
	}
	return res
}

func (ls localizedString) localize(locale string) string {
	if val, ok := ls[locale]; ok {
		return val
	} else {
		return ls[""]
	}
}

type localizedStringlist map[string][]string // Map from a locale - eg. da_DK - to a list of strings

func (lsl localizedStringlist) copy() localizedStringlist {
	var res localizedStringlist = make(localizedStringlist)
	for locale, stringlist := range lsl {
		res[locale] = utils.Copy(stringlist)
	}
	return res
}

func (ls localizedStringlist) localize(locale string) []string {
	if val, ok := ls[locale]; ok {
		return val
	} else {
		return ls[""]
	}
}
type DesktopApplication struct {
	Type            string
	Version         string
	Name            localizedString
	GenericName     localizedString
	NoDisplay       bool
	Comment         localizedString
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
	Keywords        localizedStringlist
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
	var lda =  LocalizedDesktopApplication{
		Type: da.Type,
		Version: da.Version,
		Name: da.Name.localize(locale),
		GenericName: da.GenericName.localize(locale),
		NoDisplay: da.NoDisplay,
		Comment: da.Comment.localize(locale),
		IconName: da.IconName,
		IconPath: da.IconPath,
		IconUrl: da.IconUrl,
		Hidden: da.Hidden,
		OnlyShowIn: da.OnlyShowIn,
		NotShowIn: da.NotShowIn,
		DbusActivatable: da.DbusActivatable,
		TryExec: da.TryExec,
		Exec: da.Exec,
		Path: da.Path,
		Terminal: da.Terminal,
		Mimetypes: da.Mimetypes,
		Categories: da.Categories,
		Implements: da.Implements,
		Keywords: da.Keywords.localize(locale),
		StartupNotify: da.StartupNotify,
		StartupWmClass: da.StartupWmClass,
		Url: da.Url,
		Actions: make(map[string]*LocalizedAction),
		Id: da.Id,
		RelevanceHint: da.RelevanceHint,
	}

	for id, action := range da.Actions {
		lda.Actions[id] = action.localize(locale)
	}

	return &lda
}

type Action struct {
	Name     localizedString
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
		Name: a.Name.localize(language),
		Exec: a.Exec,
		IconName: a.IconName,
		IconPath: a.IconPath,
		IconUrl: a.IconUrl,
	}
}


type IconPath string

func (ip IconPath) GET(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, string(ip))
}

func makeDesktopApplication() DesktopApplication {
	var da = DesktopApplication{
		Name: make(localizedString),
		GenericName: make(localizedString),
		Comment: make(localizedString),
		OnlyShowIn: []string{},
		NotShowIn:  []string{},
		Mimetypes:  []string{},
		Categories: []string{},
		Implements: []string{},
		Keywords:   make(localizedStringlist),
		Actions:    make(map[string]*Action),
		languages:  nil,
	}
	return da
}

func (da *DesktopApplication) Copy() *DesktopApplication {
	cp := *da
	cp.Name = cp.Name.copy()
	cp.GenericName = cp.GenericName.copy()
	cp.Comment = cp.Comment.copy()
	cp.OnlyShowIn = utils.Copy(cp.OnlyShowIn)
	cp.NotShowIn = utils.Copy(cp.NotShowIn)
	cp.Mimetypes = utils.Copy(cp.Mimetypes)
	cp.Categories = utils.Copy(cp.Categories)
	cp.Implements = utils.Copy(cp.Implements)
	cp.Keywords = cp.Keywords.copy()
	cp.Actions = make(map[string]*Action)
	for id,action := range da.Actions {
		cp.Actions[id] = action.Copy()
	}

	return &cp
}

func makeAction() Action {
	return Action{
		Name: make(localizedString),
	}
}

func (a *Action) Copy() *Action {
	copy := *a
	copy.Name = copy.Name.copy()
	return &copy
}

func (da *DesktopApplication) GET(w http.ResponseWriter, r *http.Request) {
	locale := getPreferredLocale(r, da.languages)
	resource.JsonGET(da.localize(locale), w)
}

func getPreferredLocale(r *http.Request, matcher language.Matcher) (string) {
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
	args := strings.Join(r.URL.Query()["arg"], " ")
	cmd := regexp.MustCompile("%[uUfF]").ReplaceAllString(exec, args)
	fmt.Println("Running cmd: " + cmd)
	if err := runCmd(cmd); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusAccepted)
		updatedApp := da.Copy()
		updatedApp.RelevanceHint = time.Now().UnixNano()/1000000
		service.Map(r.URL.Path, updatedApp)
	}

}

func runCmd(app string) error {
	cmd := exec.Command("sh", "-c", "("+app+">/dev/null 2>/dev/null &)")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	ioutil.ReadAll(stderr)

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

var commentLine = regexp.MustCompile(`^\s*(#.*)?$`)
func advanceToNextLine(scanner *bufio.Scanner) bool {
	for scanner.Scan() {
		if !commentLine.MatchString(scanner.Text()) {
			return true
		}
	}
	return false
}

const (
	AtStart = iota
	InDesktopEntryGroup
	InActionGroup
	InInvalidGroup
)


func readDesktopFile(path string) (*DesktopApplication, []string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return nil, nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var headerLine = regexp.MustCompile(`^\s*\[(.+?)\]\s*$`)
	var keyValueLine = regexp.MustCompile(`^\s*(.+?)(\[(.*?)\])?=(.+)`)

	var da = makeDesktopApplication()
	var actionNames = []string{}
	var currentAction = ""
	var state = 0
	var collectedLanguages = make(map[string]bool)

	for advanceToNextLine(scanner) {
		if m := headerLine.FindStringSubmatch(scanner.Text()); len(m) > 0 {
			switch state {
			case AtStart:
				if m[1] != "Desktop Entry" {
					return nil, nil, errors.New("File must start with '[Desktop Entry]'")
				} else {
					state = InDesktopEntryGroup
				}
			case InDesktopEntryGroup, InActionGroup, InInvalidGroup:
				if ! strings.HasPrefix(m[1], "Desktop Action ") {
					log.Print(path, ", ", "Unknown group type: ", m[1], "\n")
					state = InInvalidGroup
				} else if currentAction = m[1][15:]; !utils.Contains(actionNames, currentAction) {
					log.Print(path, ", undeclared action: ", currentAction, "\n")
					state = InInvalidGroup
				} else {
					var action = makeAction()
					da.Actions[currentAction] = &action
					state = InActionGroup
				}
			}
		} else if m := keyValueLine.FindStringSubmatch(scanner.Text()); len(m) > 0 {
			var tag string = transformLanguageTag(m[3])
			collectedLanguages[tag] = true

			switch state {
			case AtStart:
				return nil, nil, errors.New("File must start with '[Desktop Entry]'")
			case InDesktopEntryGroup:
				switch m[1] {
				case "Type": da.Type = m[4]
				case "Version": da.Version = m[4]
				case "Name": da.Name[tag] = m[4]
				case "GenericName": da.GenericName[tag] = m[4]
				case "NoDisplay": da.NoDisplay = m[4] == "true"
				case "Comment": da.Comment[tag] = m[4]
				case "Icon":
					if strings.HasPrefix(m[4], "/") {
						da.IconPath = m[4]
						da.IconUrl = "../icons" + m[4]
					} else {
						da.IconName = m[4]
					}
				case "Hidden": da.Hidden = m[4] == "true"
				case "OnlyShowIn": da.OnlyShowIn = utils.Split(m[4], ";")
				case "NotShowIn": da.NotShowIn = utils.Split(m[4], ";")
				case "DBusActivatable": da.DbusActivatable = m[4] == "true"
				case "TryExec": da.TryExec = m[4]
				case "Exec": da.Exec = m[4]
				case "Path": da.Path = m[4]
				case "Terminal": da.Terminal = m[4] == "true"
				case "Actions": actionNames = utils.Split(m[4], ";")
				case "MimeType": da.Mimetypes = utils.Split(m[4], ";")
				case "Categories": da.Categories = utils.Split(m[4], ";")
				case "Implements": da.Implements = utils.Split(m[4], ";")
				case "Keywords": da.Keywords[tag] = utils.Split(m[4], ";")
				case "StartupNotify": da.StartupNotify = m[4] == "true"
				case "StartupWMClass": da.StartupWmClass = m[4]
				case "URL": da.Url = m[4]
				}
			case InActionGroup:
				switch m[1] {
				case "Name": da.Actions[currentAction].Name[tag] = m[4]
				case "Icon":
					if strings.HasPrefix(m[4], "/") {
						da.Actions[currentAction].IconPath = m[4]
						da.Actions[currentAction].IconUrl = "../icons" + m[4]
					} else {
						da.Actions[currentAction].IconName = m[4]
					}
				case "Exec": da.Actions[currentAction].Exec = m[4]
				}
			}

		}
	}

	// Verify and adjust
	if da.Name[""] == "" {
		return nil, nil, errors.New("Desktop file invalid, no default 'Name' given")
	} else if da.Type == "" {
		return nil, nil, errors.New("Desktop file invalid, no 'Type' given")
	}

	for id, action := range da.Actions {
		if action.Name[""] == "" {
			return nil, nil, errors.New("Desktop file invalid, action " + id + " has no default 'Name'")
		}
		if action.IconName == "" && action.IconPath == "" {
			action.IconName = da.IconName
			action.IconPath = da.IconPath
			action.IconUrl = da.IconUrl
		}
	}

	var tags = make([]language.Tag, len(collectedLanguages))
	for locale,_ := range collectedLanguages {
		tags = append(tags, language.Make(locale))
	}
	da.languages = language.NewMatcher(tags)

	return &da, da.Mimetypes, nil
}

func transformLanguageTag(tag string) string {
	return strings.Replace(strings.Replace(tag, "_", "-", -1), "@", "-", -1)
}