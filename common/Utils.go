package common

import (
	"encoding/json"
	"strings"
	"net/http"
)

type StringSet map[string]bool

func (set *StringSet) Add(s string) {
	(*set)[s] = true
}

func (set *StringSet) Remove(s string) {
	delete(*set, s)
}

func (set *StringSet) AddAll(list []string) {
	for _, s := range list {
		set.Add(s)
	}
}

func (set *StringSet) RemoveAll(list []string) {
	for _, s := range list {
		set.Remove(s)
	}
}

func (ss StringSet) MarshalJSON() ([]byte, error) {
	keys := make([]string, 0, len(ss))
	for key := range ss {
		keys = append(keys, key)
	}
	return json.Marshal(keys)
}

func ToSet(list []string) StringSet {
	res := make(StringSet)
	for _, s := range list {
		res[s] = true
	}
	return res
}

func AppendIfNotThere(list []string, s string) []string {
	for _, v := range list {
		if v == s {
			return list
		}
	}

	return append(list, s)
}

func RemoveDublets(list []string) []string {
	seen := make(map[string]bool)
	j := 0
	for _, s := range list {
		if _, ok := seen[s]; !ok {
			list[j] = s
			j++
		}
	}
	result := make([]string, j)
	copy(result, list)
	return result
}

func Remove(list []string, element string) []string {
	result := make([]string, 0, len(list))
	for _, s := range list {
		if s != element {
			result = append(result, s)
		}
	}
	return result
}

func Split(str string, sep string) []string {
	return TrimAndFilterEmpties(strings.Split(str, sep))
}


func TrimAndFilterEmpties(stringList []string) []string {
	res := make([]string, 0)
	for _,str := range stringList {
		trimmed := strings.TrimSpace(str)
		if trimmed != "" {
			res = append(res, trimmed)
		}
	}
	return res
}

func Reverse(stringlist []string) []string {
	if  len(stringlist) <= 1 {
		return stringlist
	} else {
		return append(Reverse(stringlist[1:]), stringlist[0])
	}
}

func (ss *StringSet) Data(r *http.Request) (int, string, []byte){
	l := make([]string, 0, len(*ss))
	for s,_ := range *ss {
		l = append(l, s)
	}
	return GetJsonData(l)
}

func GetJsonData(v interface{}) (int, string, []byte){
	bytes, err := json.Marshal(v)
	if err != nil {
		panic("Could not json-marshal")
	};
	return http.StatusOK, "application/json", bytes
}

