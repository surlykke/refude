/*
 * Copyright (c) 2017 Christian Surlykke
 *
 * This file is part of the RefudeServices project.
 * It is distributed under the GPL v2 license.
 * Please refer to the GPL2 file for a copy of the license.
 */

package common

import (
	"bufio"
	"errors"
	"os"
	"regexp"
	"fmt"
)

type IniLine struct {
	Key string
	Value string
}

type IniGroup struct {
	Name  string
	Entries []IniLine
}

func (iniGroup *IniGroup) Line(key string) (IniLine, bool) {
	for _, line := range iniGroup.Entries {
		if key == line.Key {
			return line, true
		}
	}

	return IniLine{}, false
}

func (iniGroup *IniGroup) Value(key string) string {
	if line, ok := iniGroup.Line(key); ok {
		return line.Value
	} else {
		return ""
	}

}


type IniFile struct {
	Groups []IniGroup
}

func (iniFile IniFile) Group(name string) (IniGroup, bool) {
	for _, group := range iniFile.Groups {
		if name == group.Name {
			return group, true
		}
	}

	return IniGroup{}, false
}

func (iniFile IniFile) Value(groupName string, key string) string {
	for _,group := range(iniFile.Groups) {
		if groupName == group.Name {
			for _,line := range(group.Entries) {
				if key == line.Key {
					return line.Value
				}
			}
		}
	}

	return ""
}


func (iniFile IniFile) SetValue(groupName string, key string, value string) {
	var i int
	var j int
	for i = 0; i < len(iniFile.Groups); i++ {
		if groupName == iniFile.Groups[i].Name {
			break
		}
	}

	if i >= len(iniFile.Groups) {
		iniFile.Groups = append(iniFile.Groups, IniGroup{Name: groupName, Entries: make([]IniLine, 0)})
	}

	for j = 0; i < len(iniFile.Groups[i].Entries); i++ {
		if key == iniFile.Groups[i].Entries[j].Key {
			break
		}
	}

	if j >= len(iniFile.Groups[i].Entries) {
		iniFile.Groups[i].Entries = append(iniFile.Groups[i].Entries, IniLine{Key: key})
	}

	iniFile.Groups[i].Entries[j].Value = value
}

func ReadIniFile(path string) (IniFile, error) {
	var commentLine = regexp.MustCompile(`^\s*#.*`)
	var headerLine = regexp.MustCompile(`^\s*\[(.+?)\]\s*`)
	var keyValueLine = regexp.MustCompile(`^\s*(.+?)=(.+)`)

	file, err := os.Open(path)
	if err != nil {
		return IniFile{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	groups := make([]IniGroup, 0, 20)
	var currentGroup *IniGroup = nil
	for scanner.Scan() {
		if !commentLine.MatchString(scanner.Text()) {
			if m := headerLine.FindStringSubmatch(scanner.Text()); len(m) > 0 {
				groups = append(groups, IniGroup{Name: m[1], Entries: make([]IniLine, 0)})
				currentGroup = &groups[len(groups) - 1]
			} else if m := keyValueLine.FindStringSubmatch(scanner.Text()); len(m) > 0 {
				if currentGroup == nil {
					return IniFile{}, errors.New("Key value pair outside group: " + scanner.Text())
				}
				currentGroup.Entries = append(currentGroup.Entries, IniLine{Key: m[1], Value: m[2]})
			}
		}
	}
	return IniFile{Groups: groups}, nil
}


func WriteIniFile(path string, iniFile IniFile) error {
	if file, err := os.Create(path); err != nil {
		return err
	} else {
		defer file.Close()
		for _, group := range(iniFile.Groups) {
			if _, err := file.WriteString("[" + group.Name + "]\n"); err != nil {
				fmt.Println("Error writing to ", path, " ", err)
				return err
			}

			for _,line := range(group.Entries) {
				if _, err := file.WriteString(line.Key + "=" + line.Value + "\n"); err != nil {
					fmt.Println("Error writing to ", path, " ", err)
					return err
				}
			}
		}
	}

	return nil
}
