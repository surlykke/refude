// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package ini

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/surlykke/RefudeServices/lib/utils"
)

type LocalizedString map[string]string // Map from a locale - eg. da_DK - to a string

func (ls LocalizedString) Copy() LocalizedString {
	var res = make(LocalizedString, len(ls))
	for locale, str := range ls {
		res[locale] = str
	}
	return res
}

func (ls LocalizedString) LocalOrDefault(locale string) string {
	if val, ok := ls[locale]; ok {
		return val
	} else {
		return ls[""]
	}
}

type LocalizedStringlist map[string][]string // Map from a locale - eg. da_DK - to a list of strings

func (lsl LocalizedStringlist) Copy() LocalizedStringlist {
	var res = make(map[string][]string, len(lsl))
	for locale, stringlist := range lsl {
		res[locale] = utils.Copy(stringlist)
	}
	return res
}

func (ls LocalizedStringlist) LocalOrDefault(locale string) []string {
	if val, ok := ls[locale]; ok {
		return val
	} else {
		return ls[""]
	}
}

type Group struct {
	Name    string
	Entries map[string]LocalizedString
}

func (g *Group) Local(key string, locale string) string {
	if localizedString, ok := g.Entries[key]; ok {
		return localizedString[locale]
	} else {
		return ""
	}
}

type IniFile []*Group

func (inifile IniFile) FindGroup(groupName string) *Group {
	for _, group := range inifile {
		if group.Name == groupName {
			return group
		}
	}
	return nil
}

func ReadIniFile(path string) (IniFile, error) {
	var commentLine = regexp.MustCompile(`^\s*(#.*)?$`)
	var headerLine = regexp.MustCompile(`^\s*\[(.+?)\]\s*`)
	var keyValueLine = regexp.MustCompile(`^\s*(.+?)(\[(.+)\])?=(.+)`)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	var iniFile = make(IniFile, 0)
	var currentGroup *Group = nil
	for scanner.Scan() {
		if commentLine.MatchString(scanner.Text()) {
			continue
		} else if m := headerLine.FindStringSubmatch(scanner.Text()); len(m) > 0 {
			if currentGroup = iniFile.FindGroup(m[1]); currentGroup != nil {
				log.Println("Warn: iniFile", path, " has duplicate group entry: ", m[1])
			} else {
				currentGroup = &Group{m[1], make(map[string]LocalizedString)}
				iniFile = append(iniFile, currentGroup)
			}
		} else if m = keyValueLine.FindStringSubmatch(scanner.Text()); len(m) > 0 {
			var key = m[1]
			var locale = m[3]
			var value = m[4]
			if currentGroup == nil {
				return nil, errors.New("Invalid iniFile," + path + ": file must start with a group heading")
			}
			if _, ok := currentGroup.Entries[key]; !ok {
				currentGroup.Entries[key] = make(LocalizedString)
			}
			currentGroup.Entries[key][locale] = value
		} else {
			fmt.Println(scanner.Text(), " - not recognized")
		}
	}


	for _, group := range iniFile {
		for key, localizedString := range group.Entries {
			if _,ok := localizedString[""]; !ok {
				return nil, errors.New("Group " + group.Name + ", entry " + key + " has no default value")
			}
		}
	}

	return iniFile, nil
}

func WriteIniFile(path string, iniFile IniFile) error {
	if file, err := os.Create(path); err != nil {
		return err
	} else {
		defer file.Close()
		for _, group := range iniFile {
			file.WriteString("[" + group.Name + "]\n")
			for key, localizedString := range group.Entries {
				file.WriteString(key + "=" + localizedString[""] + "\n")
				for locale, value := range localizedString {
					file.WriteString(key)
					if locale != "" {
						file.WriteString("[" + locale + "]")
					}
					file.WriteString(value + "\n")
				}
			}
		}
		return nil
	}
}
