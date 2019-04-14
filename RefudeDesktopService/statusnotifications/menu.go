package statusnotifications

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/godbus/dbus"
	"github.com/pkg/errors"
	"github.com/surlykke/RefudeServices/lib/resource"
)

type Menu struct {
	resource.AbstractResource
	Menu []MenuItem
}

type MenuItem struct {
	Id          string
	Type        string
	Label       string
	Enabled     bool
	Visible     bool
	IconName    string
	Shortcuts   [][]string `json:",omitempty"`
	ToggleType  string     `json:",omitempty"`
	ToggleState int32
	SubMenus    []MenuItem `json:",omitempty"`
}

func GetMenu(path resource.StandardizedPath) *Menu {
	if item := GetItem(path); item == nil {
		return nil
	} else {
		var tmp = string(path[len("/itemmenu/"):])
		fmt.Println("tmp:", tmp)
		if slashPos := strings.Index(tmp, "/"); slashPos == -1 {
			return nil
		} else {
			var sender = tmp[0:slashPos]
			var path = tmp[slashPos:]
			fmt.Println("fetching on", sender, path)
			if menuItems, err := fetchMenu(sender, dbus.ObjectPath(path)); err != nil {
				fmt.Println("got error", err)
				return nil
			} else {
				fmt.Println("got menu")
				var menu = Menu{resource.MakeAbstractResource(resource.Standardizef("/itemmenu/%s/%s", sender, path), ""), menuItems}
				menu.LinkTo(item.GetSelf(), resource.Related)
				return &menu
			}

		}
	}
}

func (menu *Menu) POST(w http.ResponseWriter, r *http.Request) {
	/*fmt.Println("POST: ", r.URL)
	action := requests.GetSingleQueryParameter(r, "action", "left")
	x, _ := strconv.Atoi(requests.GetSingleQueryParameter(r, "x", "0"))
	y, _ := strconv.Atoi(requests.GetSingleQueryParameter(r, "y", "0"))
	id := requests.GetSingleQueryParameter(r, "id", "")

	var call *dbus.Call
	if slice.Among(action, "left", "middle", "right") {
		action2method := map[string]string{"left": "Activate", "middle": "SecondaryActivate", "right": "ContextMenu"}
		fmt.Println("Calling: ", "org.kde.StatusNotifierItem."+action2method[action], dbus.Flags(0), x, y)
		dbusObj := conn.Object(item.sender, item.itemPath)
		call = dbusObj.Call("org.kde.StatusNotifierItem."+action2method[action], dbus.Flags(0), x, y);
	} else if action == "menu" && slice.Among(id, item.menuIds...) {
		idAsInt, _ := strconv.Atoi(id)
		data := dbus.MakeVariant("")
		time := uint32(time2.Now().Unix())
		fmt.Println("Calling: ", "com.canonical.dbusmenu.Event", dbus.Flags(0), idAsInt, "clicked", data, time)
		dbusObj := conn.Object(item.sender, item.menuPath)
		call = dbusObj.Call("com.canonical.dbusmenu.Event", dbus.Flags(0), idAsInt, "clicked", data, time)
	} else {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	if call.Err != nil {
		log.Println(call.Err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusAccepted)
	}*/
}

func fetchMenu(sender string, path dbus.ObjectPath) ([]MenuItem, error) {
	obj := conn.Object(sender, path)
	fmt.Println("CALL", MENU_INTERFACE, "with", sender, path)
	if call := obj.Call(MENU_INTERFACE+".GetLayout", dbus.Flags(0), 0, -1, []string{}); call.Err != nil {
		fmt.Println("1 err:", call.Err)
		return nil, call.Err
	} else if len(call.Body) < 2 {
		fmt.Println("len(call.Body)", len(call.Body))
		return nil, errors.New(fmt.Sprint("Retrieved", len(call.Body), "arguments, expected 2"))
	} else if _, ok := call.Body[0].(uint32); !ok {
		fmt.Println("call.Body[0] not uint32")
		return nil, errors.New(fmt.Sprint("Expected uint32 as first return argument, got:", reflect.TypeOf(call.Body[0])))
	} else if interfaces, ok := call.Body[1].([]interface{}); !ok {
		fmt.Println("call.Body[1] not an interface{}")
		return nil, errors.New(fmt.Sprint("Expected []interface{} as second return argument, got:", reflect.TypeOf(call.Body[1])))
	} else if menu, err := parseMenu(interfaces); err != nil {
		fmt.Println("parse error", err)
		return nil, err
	} else if len(menu.SubMenus) > 0 {
		fmt.Println("Got several menus")
		return menu.SubMenus, nil
	} else {
		fmt.Println("Got one..")
		return []MenuItem{menu}, nil
	}
}
