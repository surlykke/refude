package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type IniGroup struct {
	name  string
	entry map[string]string
}

func ReadIniFile(path string) ([]IniGroup, error) {
	var commentLine = regexp.MustCompile(`^\s*#.*`)
	var headerLine = regexp.MustCompile(`^\s*\[(.+?)\]\s*`)
	var keyValueLine = regexp.MustCompile(`^\s*(.+?)=(.+)`)

	file, err := os.Open(path)
	if err != nil {
		return make([]IniGroup, 0), err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	iniGroups := make([]IniGroup, 0, 20)
	var currentGroup *IniGroup = nil

	for scanner.Scan() {
		if !commentLine.MatchString(scanner.Text()) {
			if m := headerLine.FindStringSubmatch(scanner.Text()); len(m) > 0 {
				iniGroups = append(iniGroups, IniGroup{})
				currentGroup = &iniGroups[len(iniGroups)-1]
				currentGroup.name = m[1]
				currentGroup.entry = make(map[string]string)
			} else if m := keyValueLine.FindStringSubmatch(scanner.Text()); len(m) > 0 {
				if currentGroup == nil {
					fmt.Println()
					return make([]IniGroup, 0), errors.New("Key value pair outside group: " + scanner.Text())
				}
				currentGroup.entry[m[1]] = m[2]
			}
		}
	}
	return iniGroups, nil
}

func split(str string) []string {
	result := make([]string, 0)
	for _, part := range strings.Split(str, ";") {
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}


