package browser

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/surlykke/refude/internal/lib/xdg"
)

var faviconCache map[string]string // url -> rawfile path
var faviconCacheLock sync.Mutex
var faviconCachePath = xdg.RefudeCacheDir + "/faviconcache.json"

func init() {
	faviconCache = make(map[string]string)
	if bytes, err := os.ReadFile(faviconCachePath); err == nil {
		json.Unmarshal(bytes, &faviconCache)
	}
}

func flushFaviconCache() {
	log.Println("Flushing faviconCache to", faviconCachePath)
	if bytes, err := json.Marshal(faviconCache); err != nil {
		log.Println("Unable to marshal faviconCache:", err)
	} else if err = os.WriteFile(faviconCachePath, bytes, 0600); err != nil {
		log.Println("Unable to write faviconCache:", err)
	}
}
