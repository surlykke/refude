package applications

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/lib/slice"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/applications" {
		respond.AsJson(w, r, CollectApps(searchutils.Term(r)))
	} else if r.URL.Path == "/mimetypes" {
		respond.AsJson(w, r, CollectMimetypes(searchutils.Term(r)))
	} else if actions := findApplicationActions(r); actions != nil {
		respond.AsJson(w, r, actions)
	} else if app := findApplication(r); app != nil {
		if r.Method == "POST" {
			respond.AcceptedAndThen(w, func() { Launch(app.Exec, app.Terminal) })
		} else {
			respond.AsJson(w, r, app.ToStandardFormat())
		}
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
	if !strings.HasPrefix(r.URL.Path, "/application/action/") {
		return nil
	} else if parts := strings.Split(r.URL.Path[20:], "/"); len(parts) != 2 {
		return nil
	} else if app := collectionStore.Load().(collection).applications[parts[0]]; app == nil {
		return nil
	} else if act := app.DesktopActions[parts[1]]; act == nil {
		return nil
	} else {
		return act
	}
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
		return "/application/action/" + appId + "/" + actionId
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

	recommended = make([]*DesktopApplication, 0, 10)
	other = make([]*DesktopApplication, 0, len(c.applications))

	for _, app := range c.applications {
		if argPlaceholders.MatchString(app.Exec) {
			for _, mt := range app.Mimetypes {
				if mt == mimetypeId {
					recommended = append(recommended, app)
					goto next
				}
			}
			other = append(other, app)
		next:
		}
	}

	return recommended, other
}
