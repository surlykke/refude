package common

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
)

type IniGroup struct {
	Name  string
	Entry map[string]string
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
				currentGroup.Name = m[1]
				currentGroup.Entry = make(map[string]string)
			} else if m := keyValueLine.FindStringSubmatch(scanner.Text()); len(m) > 0 {
				if currentGroup == nil {
					fmt.Println()
					return make([]IniGroup, 0), errors.New("Key value pair outside group: " + scanner.Text())
				}
				currentGroup.Entry[m[1]] = m[2]
			}
		}
	}
	return iniGroups, nil
}
