package monitor

import (
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/windows/x11"
)


type MonitorRepo struct {}

func (mr MonitorRepo) GetResources() []resource.Resource {
	var mdList = x11.GetMonitors()
	var resources = make([]resource.Resource, 0, len(mdList))
	for _, md := range mdList {
		resources = append(resources, md)
	}
	return resources
}

func GetMonitors() []*x11.MonitorData {
	return x11.GetMonitors()
}

func (mr MonitorRepo) GetResource(path string) resource.Resource {
	for _, md := range x11.GetMonitors() {
		if md.Path == path {
			return md
		}
	}
	return nil
}

func (mr MonitorRepo) Search(term string, threshold int) link.List {
	return link.List{}
}

var Repo MonitorRepo



