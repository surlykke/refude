package resource

type Selfie interface {
	GetSelf() string
}

type Selfielist []Selfie

func (sl Selfielist) Len() int           { return len(sl) }
func (sl Selfielist) Swap(i, j int)      { sl[i], sl[j] = sl[j], sl[i] }
func (sl Selfielist) Less(i, j int) bool { return sl[i].GetSelf() < sl[j].GetSelf() }

func (sl Selfielist) GetSelfs() []string {
	var list = make([]string, len(sl), len(sl))
	for i, sl := range sl {
		list[i] = sl.GetSelf()
	}

	return list
}
