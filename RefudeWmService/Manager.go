package main

import (
	"errors"
	"github.com/surlykke/RefudeServices/RefudeWmService/xlibutils"
	"github.com/surlykke/RefudeServices/lib"
	"log"
	"sync"
)


const (
	NET_WM_VISIBLE_NAME = "_NET_WM_VISIBLE_NAME"
	NET_WM_NAME = "_NET_WM_NAME"
	WM_NAME = "WM_NAME"
	NET_WM_ICON = "_NET_WM_ICON"
	NET_CLIENT_LIST_STACKING = "_NET_CLIENT_LIST_STACKING"
	NET_WM_STATE = "_NET_WM_STATE"
)

type Manager struct {
	in      *xlibutils.XConnection // Used to retrive data from X. Events, getProperty.. All access to this originates in the Run method
	out     *xlibutils.XConnection // Used send data to X, as events or through setters. Protected by a mutex
	outLock sync.Mutex
	windows map[uint32]*Window
}

func MakeManager() Manager {
	var m Manager
	m.in = xlibutils.MakeConnection()
	m.out = xlibutils.MakeConnection()
	m.windows = make(map[uint32]*Window)
	return m
}

func (m *Manager) findByParent(parent uint32) (*Window, bool) {
	for _, window := range m.windows {
		if window.Parent == parent {
			return window, true
		}
	}
	return nil, false
}

func (m *Manager) updateWindows() {
	if wIds, err := m.in.GetUint32s(0, NET_CLIENT_LIST_STACKING); err != nil {
		log.Fatal("Unable to get client list stacking", err)
	} else {
		var newWindows = make(map[uint32]*Window, len(wIds))
		var resourcesToMap = make([]lib.Resource, 0, 2*len(wIds))

		for i, wId := range wIds {
			if window, ok := m.windows[wId]; ok {
				window.StackOrder = -i
				newWindows[wId] = window
			} else {
				window = &Window{}
				window.Id = wId;
				window.StackOrder = -i
				window.AbstractResource = lib.AbstractResource{Self: lib.Standardizef("/window/%d", wId), Mt: WindowMediaType}
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
				window.Actions = map[string]*lib.Action2{"default": lib.MakeAction2("Raise and focus", window.IconName, executer)}
				m.windows[window.Id] = window
				resourcesToMap = append(resourcesToMap, &(*window))
				m.in.Listen(window.Id)

				if (!lib.Contains(window.States, "_NET_WM_STATE_ABOVE")) {
					window.action = lib.MakeAction(
						lib.Standardizef("/window/%d/action", wId),
						window.Name,
						"Switch to this window",
						window.IconName,
						executer)
					lib.Relate(&window.AbstractResource, &window.action.AbstractResource)
					resourcesToMap = append(resourcesToMap, window.action)
				}
			}
		}

		var prefixesToRemove = make([]lib.StandardizedPath, 0, len(wIds))

		for _, window := range m.windows {
			if _, ok := newWindows[window.Id]; !ok {
				prefixesToRemove = append(prefixesToRemove, window.Self)
			}
		}

		resourceCollection.RemoveAndMap(prefixesToRemove, resourcesToMap)
	}
}

func (m *Manager) makeRaiser(wId uint32) lib.Executer {
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
			case "":
				if window, ok := m.findByParent(event.Window); ok {
					window.X, window.Y, window.W, window.H = event.X, event.Y, event.W, event.H
					resourceCollection.Map(&(*window))
				}
			case NET_CLIENT_LIST_STACKING:
				m.updateWindows()
			case NET_WM_VISIBLE_NAME, NET_WM_NAME, WM_NAME:
				if window, ok := m.windows[event.Window]; ok {
					if window.Name, err = m.GetName(window.Id); err != nil {
						log.Println("Error getting window name:", err)
					} else {
						resourceCollection.Map(&(*window))
						if window.action != nil {
							window.action.Name = window.Name
							resourceCollection.Map(&(*(window.action)))
						}
					}
				}
			case NET_WM_ICON:
				if window, ok := m.windows[event.Window]; ok {
					if window.Name, err = m.GetIconName(window.Id); err != nil {
						log.Println("Error getting window iconname:", err)
					} else {
						resourceCollection.Map(&(*window))
						if window.action != nil {
							window.action.IconName = window.IconName
							resourceCollection.Map(&(*(window.action)))
						}
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
	/**
	  This doesn't work. openbox seems to put window titles in _NET_WM_VISIBLE_ICON_NAME and _NET_WM_ICON_NAME ??


	if name, err := xprop.PropValStr(xprop.GetProperty(xutil, wId, "_NET_WM_VISIBLE_ICON_NAME")); err == nil {
		return name
	} else if name, err := xprop.PropValStr(xprop.GetProperty(xutil, wId, "_NET_WM_ICON_NAME")); err == nil {
		return name
	} else
	*/

	iconArr, err := m.in.GetUint32s(wId, NET_WM_ICON);
	if err != nil {
			return "", err
	}

	return lib.SaveAsPngToSessionIconDir(extractARGBIcon2(iconArr)), nil
}

/**
 * Icons retrieved from the X-server (EWMH) will come as arrays of uint32. There will be first two ints giving
 * width and height, then width*height uints each holding a pixel in ARGB format.
 * After that it may repeat: again a width and height uint and then pixels and
 * so on...
 */
func extractARGBIcon2(uint32s []uint32) lib.Icon {
	res := make(lib.Icon, 0)
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
		res = append(res, lib.Img{Width: width, Height: height, Pixels: pixels})
		uint32s = uint32s[width*height:]
	}

	return res
}

