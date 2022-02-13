// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package file

import (
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

func Get(filePath string) resource.Resource {
	if file, err := makeFile(filePath); err != nil {
		log.Warn("Could not make file from", filePath, err)
		return nil
	} else if file == nil {
		return nil
	} else {
		return file
	}
}



func Search(term string) link.List {
	if file, err := makeFile(xdg.Home); err != nil {
		return link.List{}
	} else {
		return file.Related(term)
	}
}
