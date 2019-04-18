package windows

import (
	"fmt"
	"log"
	"time"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows/xlib"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/icons"
	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/resource"
)

// Returns windows in descending stack order
func getWindows() ([]*Window, error) {
	if wIds, err := xlib.GetStack(); err != nil {
		return nil, fmt.Errorf("Unable to get client list stacking %v", err)
	} else {
		var windows = make([]*Window, 0, len(wIds))
		for i := 0; i < len(wIds); i++ {
			var wId = wIds[len(wIds)-i-1]
			if window, err := getWindow(wId); err != nil {
				log.Printf("Error getting window %d: %v\n", wId, err)
			} else {
				windows = append(windows, window)
			}
		}

		return windows, nil
	}
}

func getWindow(wId uint32) (*Window, error) {
	var start = time.Now()
	window := &Window{}
	window.Id = wId
	var err error
	window.AbstractResource = resource.MakeAbstractResource(windowSelf(wId), WindowMediaType)
	window.Parent, err = xlib.GetParent(wId)
	if err != nil {
		return nil, err
	}

	if window.Parent != 0 {
		window.X, window.Y, window.W, window.H, err = xlib.GetGeometry(window.Parent)
	} else {
		window.X, window.Y, window.W, window.H, err = xlib.GetGeometry(wId)
	}
	if err != nil {
		return nil, err
	}

	window.Name, err = xlib.GetName(wId)
	if err != nil {
		return nil, err
	}

	window.IconName, err = GetIconName(wId)
	if err != nil {
		return nil, err
	}

	window.States, err = xlib.GetState(wId)
	if err != nil {
		return nil, err
	}

	var wIdCopy = wId
	var executer = func() {
		xlib.RaiseAndFocusWindow(wIdCopy)
	}

	window.ResourceActions["default"] = resource.ResourceAction{Description: "Raise and focus", IconName: window.IconName, Executer: executer}
	fmt.Println("getWindow took", time.Since(start))
	fmt.Println("")
	return window, nil
}

func GetIconName(wId uint32) (string, error) {
	pixelArray, err := xlib.GetIcon(wId)
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

	var icon = image.MakeIconWithHashAsName(images)
	icons.AddARGBIcon(icon)
	return icon.Name, nil
}
