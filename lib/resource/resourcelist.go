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

type pathAndResource struct {
	Path     string
	Resource interface{}
}

type pathAndResourceList []pathAndResource

/* sort.Interface */
func (pl pathAndResourceList) Len() int               { return len(pl) }
func (pl pathAndResourceList) Swap(i int, j int)      { pl[i], pl[j] = pl[j], pl[i] }
func (pl pathAndResourceList) Less(i int, j int) bool { return pl[i].Path < pl[j].Path }

/**
 * Returns a list of resources, sorted by path
 */
func ExtractResourceList(resources map[string]interface{}) []interface{} {
	var parl = make(pathAndResourceList, 0, len(resouces))
	for path, resource := range resources {
		parl = append(parl, pathAndResource{Path: path, Resource: resource})
	}
	sort.Sort(parl)
	var resourceList = make([]interface{}, 0, len(parl))
	for _, par := range parl {
		resourceList = append(resourceList, par.Resource)
	}

	return resourceList
}
