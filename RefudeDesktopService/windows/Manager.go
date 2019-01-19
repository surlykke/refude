package windows

import (
	"errors"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows/xlib"
	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/resource"
	"log"
	"sync"
)

const (
	NET_WM_VISIBLE_NAME      = "_NET_WM_VISIBLE_NAME"
	NET_WM_NAME              = "_NET_WM_NAME"
	WM_NAME                  = "WM_NAME"
	NET_WM_ICON              = "_NET_WM_ICON"
	NET_CLIENT_LIST_STACKING = "_NET_CLIENT_LIST_STACKING"
	NET_WM_STATE             = "_NET_WM_STATE"
)

type Manager struct {
	in             *xlib.Connection // Used to retrive data from X. Events, getProperty.. All access to this originates in the Run method
	out            *xlib.Connection // Used send data to X (events or through setters). Protected by a mutex (outLock)
	outLock        sync.Mutex
	mappedWindows  []*Window
	mappingsStream chan<- resource.Mappings
}

func Run(mappingsStream chan<- resource.Mappings) {
	var manager = Manager{
		in:             xlib.MakeConnection(),
		out:            xlib.MakeConnection(),
		mappedWindows:  []*Window{},
		mappingsStream: mappingsStream,
	}
	manager.Run()
}

func (m *Manager) mapSingle(res resource.Resource) {
	m.mappingsStream <- resource.Mappings{ResourcesToMap: map[resource.StandardizedPath]resource.Resource{res.GetSelf(): res}}
}

func (m *Manager) findIndexById(wId uint32) (int, bool) {
	for i := 0; i < len(m.mappedWindows); i++ {
		if m.mappedWindows[i].Id == wId {
			return i, true
		}
	}

	return -1, false;
}

func (m *Manager) findIndexByParentId(parentId uint32) (int, bool) {
	for i := 0; i < len(m.mappedWindows); i++ {
		if m.mappedWindows[i].Parent == parentId {
			return i, true
		}
	}
	return -1, false
}

func (m *Manager) updateWindows() {
	if wIds, err := m.in.GetUint32s(0, NET_CLIENT_LIST_STACKING); err != nil {
		log.Fatal("Unable to get client list stacking", err)
	} else {
		var newMappedWindows = make([]*Window, len(wIds))
		var resourcesToMap = make(map[resource.StandardizedPath]resource.Resource)
		for i, wId := range wIds {
			var stackOrder = len(wIds) - i
			if index, ok := m.findIndexById(wId); ok {
				var copy = *(m.mappedWindows[index])
				copy.StackOrder = stackOrder
				newMappedWindows[i] = &copy
			} else {
				window := &Window{}
				window.Id = wId;
				window.StackOrder = stackOrder
				window.AbstractResource = resource.MakeAbstractResource(resource.Standardizef("/window/%d", wId), WindowMediaType)
				if window.Parent, err = m.in.GetParent(wId); err != nil {
					log.Println("No parent:", err)
					continue
				}
				if window.X, window.Y, window.W, window.H, err = m.in.GetGeometry(wId); err != nil {
					log.Println("No geometry:", err)
					continue
				}
				if window.Name, err = m.GetName(wId); err != nil {
					log.Println("No name: ", err)
					continue
				}
				if window.IconName, err = m.GetIconName(wId); err != nil {
					log.Println("No Iconname:", err)
				}
				if window.States, err = m.in.GetAtoms(wId, NET_WM_STATE); err != nil {
					log.Println("No states: ", err)
				}

				var executer = m.makeRaiser(wId)
				window.ResourceActions["default"] = resource.ResourceAction{Description: "Raise and focus", IconName: window.IconName, Executer: executer}
				m.in.Listen(window.Id)
				newMappedWindows[i] = window
			}
			resourcesToMap[newMappedWindows[i].GetSelf()] = newMappedWindows[i]
		}

		m.mappedWindows = newMappedWindows
		m.mappingsStream <- resource.Mappings{
			PrefixesToRemove: []resource.StandardizedPath{"/windows"},
			ResourcesToMap:   resourcesToMap,
		}
	}
}

func (m *Manager) makeRaiser(wId uint32) resource.Executer {
	return func() {
		m.outLock.Lock();
		defer m.outLock.Unlock()
		m.out.RaiseAndFocusWindow(wId)
	}
}

func (m *Manager) Run() {
	m.in.Listen(0)
	m.updateWindows()

	for {
		if event, err := m.in.NextEvent(); err != nil {
			log.Println("Error retrieving next event:", err)
		} else {
			switch event.Property {
			case "": // Means it's a ConfigureEvent
				if index, ok := m.findIndexByParentId(event.Window); ok {
					var copy = *(m.mappedWindows[index])
					copy.X, copy.Y, copy.W, copy.H = event.X, event.Y, event.W, event.H
					m.mappedWindows[index] = &copy
					m.mapSingle(&copy)
				}
			case NET_CLIENT_LIST_STACKING:
				m.updateWindows()
			case NET_WM_VISIBLE_NAME, NET_WM_NAME, WM_NAME:
				if index, ok := m.findIndexById(event.Window); ok {
					var copy = *(m.mappedWindows[index])
					if copy.Name, err = m.GetName(copy.Id); err != nil {
						log.Println("Error getting copy name:", err)
					} else {
						m.mappedWindows[index] = &copy
						m.mapSingle(&copy)
					}
				}
			case NET_WM_ICON:
				if index, ok := m.findIndexById(event.Window); ok {
					var copy = *(m.mappedWindows[index])
					if copy.Name, err = m.GetIconName(copy.Id); err != nil {
						log.Println("Error getting window iconname:", err)
					} else {
						m.mappedWindows[index] = &copy
						m.mapSingle(&copy)
					}
				}
			case NET_WM_STATE:
				if index, ok := m.findIndexById(event.Window); ok {
					var copy = *(m.mappedWindows[index])
					if copy.States, err = m.in.GetAtoms(copy.Id, NET_WM_STATE); err != nil {
						log.Println("Error get window states:", err)
					} else {
						m.mappedWindows[index] = &copy
						m.mapSingle(&copy)
					}
				}
			}
		}
	}
}

func (m *Manager) GetName(wId uint32) (string, error) {
	if name, err := m.in.GetPropStr(wId, NET_WM_VISIBLE_NAME); err == nil {
		return name, nil
	} else if name, err = m.in.GetPropStr(wId, NET_WM_NAME); err == nil {
		return name, nil;
	} else if name, err = m.in.GetPropStr(wId, WM_NAME); err == nil {
		return name, nil;
	} else {
		return "", errors.New("Neither '_NET_WM_VISIBLE_NAME', '_NET_WM_NAME' nor 'WM_NAME' set")
	}
}

func (m *Manager) GetIconName(wId uint32) (string, error) {
	iconArr, err := m.in.GetUint32s(wId, NET_WM_ICON);
	if err != nil {
		return "", err
	}

	return image.SaveAsPngToSessionIconDir(extractARGBIcon2(iconArr)), nil
}

/**
 * Icons retrieved from the X-server (EWMH) will come as arrays of uint32. There will be first two ints giving
 * width and height, then width*height uints each holding a pixel in ARGB format.
 * After that it may repeat: again a width and height uint and then pixels and
 * so on...
 */
func extractARGBIcon2(uint32s []uint32) image.Icon {
	res := make(image.Icon, 0)
	for len(uint32s) >= 2 {
		width := int32(uint32s[0])
		height := int32(uint32s[1])

		uint32s = uint32s[2:]
		if len(uint32s) < int(width*height) {
			break
		}
		pixels := make([]byte, 4*width*height)
		for pos := int32(0); pos < width*height; pos++ {
			pixels[4*pos] = uint8((uint32s[pos] & 0xFF000000) >> 24)
			pixels[4*pos+1] = uint8((uint32s[pos] & 0xFF0000) >> 16)
			pixels[4*pos+2] = uint8((uint32s[pos] & 0xFF00) >> 8)
			pixels[4*pos+3] = uint8(uint32s[pos] & 0xFF)
		}
		res = append(res, image.Img{Width: width, Height: height, Pixels: pixels})
		uint32s = uint32s[width*height:]
	}

	return res
}
