// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
package browser

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/surlykke/refude/internal/lib/xdg"
	"github.com/surlykke/refude/internal/rawfile"
)

var faviconCache map[string]string // url -> rawfile path
var faviconCacheLock sync.Mutex
var faviconCachePath = xdg.RefudeCacheDir + "/faviconcache.json"

func init() {
	faviconCache = make(map[string]string)
	if bytes, err := os.ReadFile(faviconCachePath); err == nil {
		if json.Unmarshal(bytes, &faviconCache) == nil {
			var somethingDeleted = false
			for k, v := range faviconCache {
				if strings.HasPrefix(v, rawfile.Prefix+"/") {
					rawfile.Register(v[len(rawfile.Prefix):])
				} else {
					delete(faviconCache, k)
					somethingDeleted = true
				}
			}
			if somethingDeleted {
				flushFaviconCache()
			}
		}
	}
}

func flushFaviconCache() {
	if bytes, err := json.Marshal(faviconCache); err != nil {
		log.Println("Unable to marshal faviconCache:", err)
	} else if err = os.WriteFile(faviconCachePath, bytes, 0600); err != nil {
		log.Println("Unable to write faviconCache:", err)
	}
}
