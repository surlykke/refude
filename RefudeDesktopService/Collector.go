package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/surlykke/RefudeServices/lib/utils"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

type association struct {
	mimetype    string
	application string
}

type collection struct {
	mimetypes    map[string]*Mimetype           // Maps from mimetypeid to mimetype
	applications map[string]*DesktopApplication // Maps from applicationid to application
	defaultApps  map[string][]string            // Maps from mimetypeid to list of app ids
}

func Collect() collection {
	var c collection
	c.mimetypes = CollectMimeTypes()
	c.applications = make(map[string]*DesktopApplication)
	c.defaultApps = make(map[string][]string)

	for _, dir := range xdg.DataDirs {
		c.collectApplications(dir + "/applications")
		c.readMimeappsList(dir + "/applications/mimeapps.list")
	}

	c.collectApplications(xdg.DataHome + "/applications")

	for _, dir := range append(xdg.ConfigDirs, xdg.ConfigHome) {
		c.readMimeappsList(dir + "/mimeapps.list")
	}

	for appId, app := range c.applications {
		for _, mimetypeId := range app.Mimetypes {
			if mimetype, ok := c.mimetypes[mimetypeId]; ok {
				utils.AppendIfNotThere(mimetype.AssociatedApplications, appId)
			}
		}
	}

	for mimetypeId, appIds := range c.defaultApps {
		if mimetype, ok := c.mimetypes[mimetypeId]; ok {
			for _, appId := range appIds {
				if _, ok := c.applications[appId]; ok {
					mimetype.DefaultApplication = appId
					break
				}
			}
		}
	}

	return c
}

func (c *collection) collectApplications(appdir string) {
	filepath.Walk(appdir, func(path string, info os.FileInfo, err error) error {
		if !(info.IsDir() || !strings.HasSuffix(path, ".desktop")) {
			app, err := readDesktopFile(path)
			if err == nil {
				app.Id = strings.Replace(path[len(appdir)+1:], "/", "-", -1)
				if app.Hidden ||
					(len(app.OnlyShowIn) > 0 && !utils.ElementsInCommon(xdg.CurrentDesktop, app.OnlyShowIn)) ||
					(len(app.NotShowIn) > 0 && utils.ElementsInCommon(xdg.CurrentDesktop, app.NotShowIn)) {
					delete(c.applications, app.Id)
				} else {
					c.applications[app.Id] = app
				}
			} else {
				log.Println("Error processing ", path, ":\n\t", err)
			}
		}
		return nil
	})

}

type mimeAppsListReadState int

const (
	AtStart mimeAppsListReadState = iota
	InDefaultAppsGroup
	InAddedAssociationsGroup
	InRemovedAssociationsGroup
	InUnknownGroup
)

var groupHeading = regexp.MustCompile(`^\s*\[(.*?)\]\s*$`)
var keyValueLine = regexp.MustCompile(`^\s*(.+?)=(.+)`)
var commentLine = regexp.MustCompile(`^\s*(#.*)?$`)

func (c *collection) readMimeappsList(path string) {
	file, err := os.Open(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Println("Error reading", path, ":", err)
		}
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var state = AtStart

	for scanner.Scan() {
		if m := groupHeading.FindStringSubmatch(scanner.Text()); len(m) > 0 {
			switch m[1] {
			case "Added Associations":
				state = InAddedAssociationsGroup
			case "Removed Associations":
				state = InRemovedAssociationsGroup
			case "Default Applications":
				state = InDefaultAppsGroup
			default:
				state = InUnknownGroup
			}
		} else if m = keyValueLine.FindStringSubmatch(scanner.Text()); len(m) > 0 {
			var mimetypeId = m[1]
			var applicationIds = utils.Split(m[2], ";")
			switch state {
			case InAddedAssociationsGroup:
				for _, appId := range applicationIds {
					if app, ok := c.applications[appId]; ok {
						utils.AppendIfNotThere(app.Mimetypes, mimetypeId)
					}
				}
			case InRemovedAssociationsGroup:
				for _, appId := range applicationIds {
					if app, ok := c.applications[appId]; ok {
						utils.Remove(app.Mimetypes, mimetypeId)
					}
				}
			case InDefaultAppsGroup:
				c.defaultApps[mimetypeId] = append(applicationIds, c.defaultApps[mimetypeId]...)
			}
		} else if m = commentLine.FindStringSubmatch(scanner.Text()); len(m) == 0 {
			log.Println("Skipping incomprehensible line: ", scanner.Text())
		}
	}
}
