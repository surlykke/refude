package backlight

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/surlykke/RefudeServices/lib/resource"
)

const backlightdir = "/sys/class/backlight"
const brightness = "/brightness"
const lenbld = len(backlightdir)
const lenbn = len(brightness)

type device struct {
	Id            string
	BrightnessPct uint8
	maxBrightness uint64
	brightness    uint64
}

func (dev *device) setBrightness(brightness uint64) {
	if brightness > dev.maxBrightness {
		dev.brightness = 0
	} else {
		dev.brightness = brightness
	}

	dev.BrightnessPct = uint8(100 * dev.brightness / dev.maxBrightness)
}

var lock = &sync.Mutex{}
var devices = make(map[string]*device)

var watcher = func() *fsnotify.Watcher {
	if tmp, err := fsnotify.NewWatcher(); err != nil {
		panic(err)
	} else if err := tmp.Add(backlightdir); err != nil {
		panic(fmt.Errorf("Unable to watch %s: %s", backlightdir, err))
	} else {
		return tmp
	}
}()

func getDevices() {
	for id, _ := range devices {
		watcher.Remove(backlightdir + "/" + id + "/brightness")
	}
	if dirs, err := filepath.Glob(backlightdir + "/*"); err != nil {
		panic(fmt.Sprint("Error globbing for backlight catalogues: ", err))
	} else {
		for _, dir := range dirs {
			var dev = &device{Id: dir[len(backlightdir)+1:]}
			var maxBrightnessPath = dir + "/max_brightness"
			var brightnessPath = dir + "/brightness"
			if err := watcher.Add(brightnessPath); err != nil {
				fmt.Println("Unable to watch", brightnessPath, err)
			} else if dev.maxBrightness, err = readUint64(maxBrightnessPath); err != nil {
				fmt.Println("Unable to read", maxBrightnessPath, err)
				watcher.Remove(brightnessPath)
			} else if dev.brightness, err = readBrightness(brightnessPath, dev.maxBrightness); err != nil {
				fmt.Println("Problem reading brightness:", brightnessPath, err)
				watcher.Remove(brightnessPath)
			}
			dev.BrightnessPct = uint8(100 * dev.brightness / dev.maxBrightness)
			devices[dev.Id] = dev
		}
	}

	var resources = make(map[string]interface{})
	for _, dev := range devices {
		resources["/backlight/"+dev.Id] = &(*dev)
	}
	var backlightpaths, backlights = resource.ExtractPathAndResourceLists(resources)
	resources["/backlightpaths"] = backlightpaths
	resources["/backlights"] = backlights
	resource.MapCollection(&resources, "backlights")
}

func readBrightness(brightnessPath string, maxBrightness uint64) (uint64, error) {
	if brightness, err := readUint64(brightnessPath); err != nil {
		return 0, err
	} else if brightness > maxBrightness {
		return 0, fmt.Errorf("brightness > max_brightnes: %d %d", brightness, maxBrightness)
	} else {
		return brightness, nil
	}
}

func readUint64(filepath string) (uint64, error) {
	if bytes, err := ioutil.ReadFile(filepath); err != nil {
		return 0, err
	} else if val, err := strconv.ParseUint(strings.TrimSpace(string(bytes)), 10, 64); err != nil {
		return 0, err
	} else {
		return val, nil
	}
}

func Run() {
	getDevices()
	for {
		select {
		case _ = <-watcher.Events:
			getDevices()
		case err := <-watcher.Errors:
			fmt.Println("Error from watcher:", err)
		}
	}
}
