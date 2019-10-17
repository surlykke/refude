package resource

type ResourceList []Resource

/* sort.Interface */
func (rl ResourceList) Len() int           { return len(rl) }
func (rl ResourceList) Swap(i, j int)      { rl[i], rl[j] = rl[j], rl[i] }
func (rl ResourceList) Less(i, j int) bool { return rl[i].GetSelf() < rl[j].GetSelf() }

/* resource.Resource */
func (ResourceList) GetSelf() string { return "" }

type PathList []string

/* sort.Interface */
func (pl PathList) Len() int               { return len(pl) }
func (pl PathList) Swap(i int, j int)      { pl[i], pl[j] = pl[j], pl[i] }
func (pl PathList) Less(i int, j int) bool { return pl[i] < pl[j] }

/* resource.Resource */
func (bl PathList) GetSelf() string { return "" }
