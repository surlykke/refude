package windows

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows/xlib"
	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/resource"
)

const (
	NET_WM_VISIBLE_NAME      = "_NET_WM_VISIBLE_NAME"
	NET_WM_NAME              = "_NET_WM_NAME"
	WM_NAME                  = "WM_NAME"
	NET_WM_ICON              = "_NET_WM_ICON"
	NET_CLIENT_LIST_STACKING = "_NET_CLIENT_LIST_STACKING"
	NET_WM_STATE             = "_NET_WM_STATE"
)

var in = xlib.MakeConnection()  // Used to retrive data from X. Events, getProperty.. All access to this originates in the Run function
var out = xlib.MakeConnection() // Used send data to X (events or through setters). Protected by a mutex (outLock)
var outLock sync.Mutex

func handle(event xlib.Event) {
	switch event.Property {
	case NET_CLIENT_LIST_STACKING:
		updateWindows()
	case "": // Means it's a ConfigureEvent
		if copy := getCopyByParent(event.Window); copy != nil {
			copy.X, copy.Y, copy.W, copy.H = event.X, event.Y, event.W, event.H
			setWindow(copy)
		}
	case NET_WM_VISIBLE_NAME, NET_WM_NAME, WM_NAME:
		if name, err := GetName(event.Window); err != nil {
			log.Println("Error getting window name:", err)
		} else {
			if copy := GetCopy(event.Window); copy != nil {
				copy.Name = name
				setWindow(copy)
			}
		}
	case NET_WM_ICON:
		if iconName, err := GetIconName(event.Window); err != nil {
			log.Println("Error getting window iconname:", err)
		} else {
			if copy := GetCopy(event.Window); copy != nil {
				copy.IconName = iconName
				setWindow(copy)
			}
		}
	case NET_WM_STATE:
		if states, err := in.GetAtoms(event.Window, NET_WM_STATE); err != nil {
			log.Println("Error get window states:", err)
		} else {
			if copy := GetCopy(event.Window); copy != nil {
				copy.States = states
				setWindow(copy)
			}
		}
	}
}

func updateWindows() {
	if wIds, err := in.GetUint32s(0, NET_CLIENT_LIST_STACKING); err != nil {
		log.Fatal("Unable to get client list stacking", err)
	} else {
		ClearAll()
		fmt.Println("her, wIds:", wIds)
		for i, wId := range wIds {
			fmt.Println("wId:", wId)
			var stackOrder = len(wIds) - i
			window := &Window{}
			window.Id = wId
			window.StackOrder = stackOrder
			window.AbstractResource = resource.MakeAbstractResource(windowSelf(wId), WindowMediaType)
			fmt.Println("GetParent")
			if window.Parent, err = in.GetParent(wId); err != nil {
				log.Println("No parent:", err)
				continue
			}
			fmt.Println("GetGeometry")
			if window.X, window.Y, window.W, window.H, err = in.GetGeometry(wId); err != nil {
				log.Println("No geometry:", err)
				continue
			}
			fmt.Println("GetName")
			if window.Name, err = GetName(wId); err != nil {
				log.Println("No name: ", err)
				continue
			}
			fmt.Println("GetIconName")
			if window.IconName, err = GetIconName(wId); err != nil {
				log.Println("No Iconname:", err)
			}
			fmt.Println("GetState")
			if window.States, err = in.GetAtoms(wId, NET_WM_STATE); err != nil {
				log.Println("No states: ", err)
			}

			var wIdCopy = wId
			var executer = func() {
				outLock.Lock()
				defer outLock.Unlock()
				out.RaiseAndFocusWindow(wIdCopy)
			}

			window.ResourceActions["default"] = resource.ResourceAction{Description: "Raise and focus", IconName: window.IconName, Executer: executer}
			fmt.Println("listen")
			in.Listen(window.Id)
			fmt.Println("SetWindow")
			setWindow(window)
			fmt.Println("End of loop")
		}
	}
}

func GetName(wId uint32) (string, error) {
	if name, err := in.GetPropStr(wId, NET_WM_VISIBLE_NAME); err == nil {
		return name, nil
	} else if name, err = in.GetPropStr(wId, NET_WM_NAME); err == nil {
		return name, nil
	} else if name, err = in.GetPropStr(wId, WM_NAME); err == nil {
		return name, nil
	} else {
		return "", errors.New("Neither '_NET_WM_VISIBLE_NAME', '_NET_WM_NAME' nor 'WM_NAME' set")
	}
}

func GetIconName(wId uint32) (string, error) {
	fmt.Println("Getting icon name")
	pixelArray, err := in.GetUint32s(wId, NET_WM_ICON)
	if err != nil {
		return "", err
	}

	/*
	 * Icons retrieved from the X-server (EWMH) will come as arrays of uint32. There will be first two ints giving
	 * width and height, then width*height uints each holding a pixel in ARGB format.
	 * After that it may repeat: again a width and height uint and then pixels and
	 * so on...
	 */
	var images = []image.ARGBImage{}
	for len(pixelArray) >= 2 {
		width := pixelArray[0]
		height := pixelArray[1]

		pixelArray = pixelArray[2:]
		if len(pixelArray) < int(width*height) {
			break
		}
		pixels := make([]byte, 4*width*height)
		for pos := uint32(0); pos < width*height; pos++ {
			pixels[4*pos] = uint8((pixelArray[pos] & 0xFF000000) >> 24)
			pixels[4*pos+1] = uint8((pixelArray[pos] & 0xFF0000) >> 16)
			pixels[4*pos+2] = uint8((pixelArray[pos] & 0xFF00) >> 8)
			pixels[4*pos+3] = uint8(pixelArray[pos] & 0xFF)
		}
		images = append(images, image.ARGBImage{Width: width, Height: height, Pixels: pixels})
		pixelArray = pixelArray[width*height:]
	}

	fmt.Println("MakeIconWithHashAsName")
	var icon = image.MakeIconWithHashAsName(images)
	//icons.AddARGBIcon(icon)
	return icon.Name, nil
}
