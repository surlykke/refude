package network

import (
	"fmt"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/surlykke/refude/internal/lib/entity"
	"github.com/surlykke/refude/internal/lib/utils"
)

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

var Connections = entity.MakeMap[dbus.ObjectPath, *Connection]("/network/")
var primaryConnection = dbus.ObjectPath("")
var _signals = make(chan *dbus.Signal, 50)

type Connection struct {
	entity.Base
	Primary    bool
	Vpn        bool
	Type       string
	ObjectPath dbus.ObjectPath
}

/*func (this *Connection) DoPost(action string) bind.Response {
	fmt.Println(">>>>>>>>>>>>>>>>> DoPost")
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
	var c = &Connection{Base: *entity.MakeBase(title, "", "", ""), Primary: primary, Vpn: vpn, ObjectPath: path}
	var actionName string
	if primary {
		actionName = "Deactivate"
	} else {
		actionName = "Activate"
	}
	c.AddAction("", actionName, "")
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
		fmt.Println(err)
	} else {
		setPrimaryConnection(primaryConnection)
	}

	if activeConnectionPaths, err := utils.Prop[[]dbus.ObjectPath](conn(), SERVICE, NM, NMI, "ActiveConnections"); err != nil {
		fmt.Println(err)
	} else {
		setActiveConnections(activeConnectionPaths)
	}
}

func watch() {
	fmt.Println("Watching")
	for s := range _signals {
		if s.Name == utils.PROPERTIES_INTERFACE+".PropertiesChanged" && s.Path == NM {
			fmt.Println("Got signal:", s)
			if activeConnectionsVariant, ok := s.Body[1].(map[string]dbus.Variant)["ActiveConnections"]; ok {
				setActiveConnections(activeConnectionsVariant.Value().([]dbus.ObjectPath))
			} else if primaryConnectionVariant, ok := s.Body[1].(map[string]dbus.Variant)["PrimaryConnection"]; ok {
				setPrimaryConnection(primaryConnectionVariant.Value().(dbus.ObjectPath))
			}
		}
	}
}

func setActiveConnections(acPaths []dbus.ObjectPath) {
	var activeConnections = make(map[dbus.ObjectPath]*Connection)
	for _, acPath := range acPaths {
		fmt.Println("Look at", acPath, NMI+".Connection.Active")
		if props, err := utils.Props(conn(), SERVICE, acPath, NMI+".Connection.Active"); err != nil {
			fmt.Println(err)
		} else {
			var vpn = props["Vpn"] != nil && props["Vpn"].(bool)
			var Type = props["Type"].(string)
			fmt.Println("Type:", Type)
			activeConnections[acPath[1:]] = makeConnection(props["Id"].(string), primaryConnection == acPath, vpn, Type, acPath)
		}
	}
	Connections.ReplaceAll(activeConnections)
	fmt.Println("connections now:")
	for _, conn := range Connections.GetAll() {
		fmt.Println(conn)
	}
}

func setPrimaryConnection(connection dbus.ObjectPath) {
	fmt.Println("Set primary connection: ", connection)
	if connection == "/" {
		primaryConnection = ""
	} else {
		primaryConnection = connection
	}
	var copies = make(map[dbus.ObjectPath]*Connection)
	for _, conn := range Connections.GetAll() {
		var copy = *conn
		copy.Primary = copy.ObjectPath == primaryConnection
		copies[copy.ObjectPath] = &copy
	}
	Connections.ReplaceAll(copies)
}

func Run() {
	Connections.Serve()
	subscribe()
	initialize()
	watch()
}
