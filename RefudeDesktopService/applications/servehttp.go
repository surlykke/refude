package applications

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/lib/slice"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if "/applications" == r.URL.Path {
		respond.AsJson(w, r, CollectApps(searchutils.Term(r)))
	} else if "/mimetypes" == r.URL.Path {
		respond.AsJson(w, r, CollectMimetypes(searchutils.Term(r)))
	} else if app := findApplication(r); app != nil {
		if r.Method == "POST" {
			respond.AcceptedAndThen(w, func() { Launch(app.Exec, app.Terminal) })
		} else {
			respond.AsJson(w, r, app.ToStandardFormat())
		}
	} else if actions := findApplicationActions(r); actions != nil {
		respond.AsJson(w, r, actions)
	} else if act := findApplicationAction(r); act != nil {
		if r.Method == "POST" {
			respond.AcceptedAndThen(w, func() { Launch(act.Exec, false) })
		} else {
			respond.AsJson(w, r, act.ToStandardFormat())
		}
	} else if mt := findMimetype(r); mt != nil {
		respond.AsJson(w, r, mt.ToStandardFormat())
	} else {
		respond.NotFound(w)
	}

}

func findApplication(r *http.Request) *DesktopApplication {
	if strings.HasPrefix(r.URL.Path, "/application/") {
		if app, ok := collectionStore.Load().(collection).applications[r.URL.Path[13:]]; ok {
			return app
		}
	}
	return nil
}

func findApplicationActions(r *http.Request) respond.StandardFormatList {
	if strings.HasPrefix(r.URL.Path, "/application/actions/") {
		if app, ok := collectionStore.Load().(collection).applications[r.URL.Path[21:]]; ok {
			if actions := app.collectActions(searchutils.Term(r)); len(actions) > 0 {
				return actions
			}
		}
	}
	return nil
}

func findApplicationAction(r *http.Request) *DesktopAction {
	if strings.HasPrefix(r.URL.Path, "/application/action/") {
		if app, ok := collectionStore.Load().(collection).applications[r.URL.Path[20:]]; ok {
			if actionId := requests.GetSingleQueryParameter(r, "action", ""); actionId != "" {
				if action, ok := app.DesktopActions[actionId]; ok {
					return action
				}
			}
		}
	}
	return nil
}

func findMimetype(r *http.Request) *Mimetype {
	if strings.HasPrefix(r.URL.Path, "/mimetype/") {
		if mt, ok := collectionStore.Load().(collection).mimetypes[r.URL.Path[10:]]; ok {
			return mt
		}
	}

	return nil
}

func appSelf(appId string) string {
	return "/application/" + appId
}

func otherActionsPath(appId string) string {
	if !strings.HasSuffix(appId, ".desktop") {
		log.Println("Weird application id:", appId)
		return ""
	} else {
		return "/application/actions/" + appId
	}
}

func actionPath(appId, actionId string) string {
	if !strings.HasSuffix(appId, ".desktop") {
		log.Println("Weird application id:", appId)
		return ""
	} else {
		return "/application/action/" + appId + "?action=" + actionId
	}
}

func mimetypeSelf(mimetypeId string) string {
	return fmt.Sprintf("/mimetype/%s", mimetypeId)
}

func CollectApps(term string) respond.StandardFormatList {
	var c = collectionStore.Load().(collection)
	var sfl = make(respond.StandardFormatList, 0, len(c.applications))
	for _, app := range c.applications {
		if rank := searchutils.SimpleRank(app.Name, app.Comment, term); rank > -1 {
			sfl = append(sfl, app.ToStandardFormat().Ranked(rank))
		}
	}
	return sfl.SortByRank()
}

func CollectMimetypes(term string) respond.StandardFormatList {
	var c = collectionStore.Load().(collection)
	var sfl = make(respond.StandardFormatList, 0, len(c.mimetypes))
	for _, mt := range c.mimetypes {
		if rank := searchutils.SimpleRank(mt.Comment, mt.Acronym, term); rank > -1 {
			sfl = append(sfl, mt.ToStandardFormat().Ranked(rank))
		}
	}
	return sfl.SortByRank()
}

func AllPaths() []string {
	var c = collectionStore.Load().(collection)
	var paths = make([]string, 0, len(c.applications)+len(c.mimetypes)+100)
	for appId, app := range c.applications {
		paths = append(paths, appSelf(appId))
		if len(app.DesktopActions) > 0 {
			paths = append(paths, otherActionsPath(appId))
			for _, act := range app.DesktopActions {
				paths = append(paths, act.self)
			}
		}
	}
	for mimetypeId := range c.mimetypes {
		paths = append(paths, mimetypeSelf(mimetypeId))
	}
	paths = append(paths, "/applications", "/mimetypes")
	return paths
}

func GetMtList(mimetypeId string) []string {
	var mimetypes = collectionStore.Load().(collection).mimetypes
	var result = make([]string, 1, 5)
	result[0] = mimetypeId
	for i := 0; i < len(result); i++ {
		if mt, ok := mimetypes[result[i]]; ok {
			for _, super := range mt.SubClassOf {
				result = slice.AppendIfNotThere(result, super)
			}
		}
	}
	return result
}

func GetApp(appId string) *DesktopApplication {
	return collectionStore.Load().(collection).applications[appId]
}

func GetAppsForMimetype(mimetypeId string) (recommended, other []*DesktopApplication) {
	var c = collectionStore.Load().(collection)
	var handled = make(map[string]bool)
	recommended = make([]*DesktopApplication, 0, 10)
	other = make([]*DesktopApplication, 0, len(c.applications))

	for _, mimetypeId := range GetMtList(mimetypeId) {
		for _, appId := range c.associations[mimetypeId] {
			if !handled[appId] {
				if da, ok := c.applications[appId]; ok {
					recommended = append(recommended, da)
				}
				handled[appId] = true
			}
		}
	}

	for appId, app := range c.applications {
		if !handled[appId] {
			other = append(other, app)
			handled[appId] = true
		}
	}

	return recommended, other
}
