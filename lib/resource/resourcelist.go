// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package resource

import (
	"sort"
)

type PathList []string

/* sort.Interface */
func (pl PathList) Len() int               { return len(pl) }
func (pl PathList) Swap(i int, j int)      { pl[i], pl[j] = pl[j], pl[i] }
func (pl PathList) Less(i int, j int) bool { return pl[i] < pl[j] }

/**
 * Returns a list of paths and a list of resources, both sorted by path
 */
func ExtractPathAndResourceLists(resources map[string]interface{}) ([]string, []interface{}) {
	var pathList = make([]string, 0, len(resouces))
	var resourceList = make([]interface{}, 0, len(resources))
	for path, _ := range resources {
		pathList = append(pathList, path)
	}
	sort.Sort(PathList(pathList))
	for _, path := range pathList {
		resourceList = append(resourceList, resources[path])
	}

	return pathList, resourceList
}
