// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
//
package network

import (
	"log"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/surlykke/refude/internal/lib/resource"
	"github.com/surlykke/refude/internal/lib/utils"
)

const networkPathPrefix = "/network/"

const SERVICE = "org.freedesktop.NetworkManager"
const NM = "/org/freedesktop/NetworkManager"
const NMI = "org.freedesktop.NetworkManager"

var _conn *dbus.Conn
var _connM sync.Mutex

func conn() *dbus.Conn {
	_connM.Lock()
	defer _connM.Unlock()

	var err error
	if _conn == nil {
		if _conn, err = dbus.SystemBus(); err != nil {
			panic(err)
		}
	}
	return _conn
}

var primaryConnection = dbus.ObjectPath("")
var _signals = make(chan *dbus.Signal, 50)

type Connection struct {
	resource.Base
	Primary    bool
	Vpn        bool
	Type       string
	ObjectPath dbus.ObjectPath
}

/*func (this *Connection) DoPost(action string) bind.Response {
	return bind.Accepted()
	/*var dbusMember string
	if this.Primary {
		dbusMember = "ActivateConnection"
	} else {
		dbusMember = "DeactivateConnection"
	}

	if action == "" {
		if _, err := utils.Call[dbus.ObjectPath](conn(), SERVICE, dbus.ObjectPath(NM), dbusMember, this.ObjectPath, "", ""); err != nil {
			return bind.ServerError(err)
		} else {
			return bind.Accepted()
		}
	} else {
		return bind.NotFound()
	}
}*/

func makeConnection(title string, primary bool, vpn bool, connType string, path dbus.ObjectPath) *Connection {
	var c = &Connection{Base: *resource.MakeBase(networkPathPrefix+string(path), title, "Network connection", "", ""), Primary: primary, Vpn: vpn, ObjectPath: path}
	var actionName string
	if primary {
		actionName = "Deactivate"
	} else {
		actionName = "Activate"
	}
	c.AddAction(actionName, "", resource.OrgRefudeAction, "")
	c.Keywords = []string{"network"}
	return c
}

func subscribe() {
	conn().Signal(_signals)
	conn().AddMatchSignal(dbus.WithMatchSender(NM))

	//conn().AddMatchSignal()
}

func initialize() {
	if primaryConnection, err := utils.Prop[dbus.ObjectPath](conn(), SERVICE, NM, NMI, "PrimaryConnection"); err != nil {
		log.Println(err)
	} else {
		setPrimaryConnection(primaryConnection)
	}

	if activeConnectionPaths, err := utils.Prop[[]dbus.ObjectPath](conn(), SERVICE, NM, NMI, "ActiveConnections"); err != nil {
		log.Println(err)
	} else {
		setActiveConnections(activeConnectionPaths)
	}
}

func watch() {
	for s := range _signals {
		if s.Name == utils.PROPERTIES_INTERFACE+".PropertiesChanged" && s.Path == NM {
			if activeConnectionsVariant, ok := s.Body[1].(map[string]dbus.Variant)["ActiveConnections"]; ok {
				setActiveConnections(activeConnectionsVariant.Value().([]dbus.ObjectPath))
			} else if primaryConnectionVariant, ok := s.Body[1].(map[string]dbus.Variant)["PrimaryConnection"]; ok {
				setPrimaryConnection(primaryConnectionVariant.Value().(dbus.ObjectPath))
			}
		}
	}
}

func setActiveConnections(acPaths []dbus.ObjectPath) {
	var connections []*Connection
	for _, acPath := range acPaths {
		if props, err := utils.Props(conn(), SERVICE, acPath, NMI+".Connection.Active"); err != nil {
			log.Println(err)
		} else {
			var vpn = props["Vpn"] != nil && props["Vpn"].(bool)
			var Type = props["Type"].(string)
			connections = append(connections, makeConnection(props["Id"].(string), primaryConnection == acPath, vpn, Type, acPath))
		}
	}
	resource.Replace(connections, "/network/")
}

func setPrimaryConnection(connection dbus.ObjectPath) {
	if connection == "/" {
		primaryConnection = ""
	} else {
		primaryConnection = connection
	}
	var copies []*Connection
	for _, conn := range resource.GetByPrefix[*Connection]("/network/") {
		var copy = *conn
		copy.Primary = copy.ObjectPath == primaryConnection
		copies = append(copies, &copy)
	}
	resource.Replace(copies, "/network/")
}

func Run() {
	subscribe()
	initialize()
	watch()
}
