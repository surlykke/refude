package resource

import "net/http"

type ResourceList []Resource

/* sort.Interface */
func (rl ResourceList) Len() int           { return len(rl) }
func (rl ResourceList) Swap(i, j int)      { rl[i], rl[j] = rl[j], rl[i] }
func (rl ResourceList) Less(i, j int) bool { return rl[i].GetSelf() < rl[j].GetSelf() }

/* resource.Resource */
func (ResourceList) GetSelf() string { return "" }
func (ResourceList) POST(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}
func (ResourceList) PATCH(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}
func (ResourceList) DELETE(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

type BriefList []string

/* sort.Interface */
func (bl BriefList) Len() int               { return len(bl) }
func (bl BriefList) Swap(i int, j int)      { bl[i], bl[j] = bl[j], bl[i] }
func (bl BriefList) Less(i int, j int) bool { return bl[i] < bl[j] }

/* resource.Resource */
func (bl BriefList) GetSelf() string { return "" }
func (BriefList) POST(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}
func (BriefList) PATCH(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}
func (BriefList) DELETE(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}
