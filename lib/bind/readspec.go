package bind

import (
	"cmp"
	"net/http"
	"regexp"
	"strings"

	"github.com/surlykke/RefudeServices/lib/entity"
	"github.com/surlykke/RefudeServices/lib/repo"
)

func ServeFunc(path string, function any, bindings ...string) {
	http.HandleFunc(path, buildHandler(function, readSpecs(bindings...)))
}

func ServeMap[K cmp.Ordered, V entity.Servable](prefix string, m *repo.SyncMap[K, V]) {
	m.SetPrefix(prefix)
	ServeFunc("GET "+prefix+"{id...}", m.DoGetSingle, "path id")
	ServeFunc("GET "+prefix+"{$}", m.DoGetAll)
	ServeFunc("POST "+prefix+"{id...}", m.DoPost, "path id", "query action")
}

var queryWithDefault = regexp.MustCompile(`^query (\w+) default (\S+)$`)
var queryRequired = regexp.MustCompile(`^query (\w+) required$`)
var query = regexp.MustCompile(`^query (\w+)$`)
var path = regexp.MustCompile(`^path (\w+)$`)
var body = regexp.MustCompile(`^body json$`)
var errmsg = `spec must be on one of the forms:
	query <parmname> default <def-val>
	query <parmname> required 
	query <parmname> 
	path <parmname> 
	body <body-type>

With only body-type currently supported 'json'
`

func readSpecs(specs ...string) []binding {
	var bindings = make([]binding, len(specs), len(specs))
	for i, spec := range specs {
		spec = strings.Join(strings.Fields(spec), " ") // Normalize whitespace

		if m := queryWithDefault.FindStringSubmatch(spec); m != nil {
			bindings[i] = binding{query: m[1], def: m[2]}
		} else if m := queryRequired.FindStringSubmatch(spec); m != nil {
			bindings[i] = binding{query: m[1], req: true}
		} else if m := query.FindStringSubmatch(spec); m != nil {
			bindings[i] = binding{query: m[1]}
		} else if m := path.FindStringSubmatch(spec); m != nil {
			bindings[i] = binding{path: m[1]}
		} else if m := body.FindStringSubmatch(spec); m != nil {
			bindings[i] = binding{body: "json"}
		} else {
			panic("Error reading '" + spec + "'\n" + errmsg)
		}
	}
	return bindings
}
