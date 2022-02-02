package relation

type Relation uint8

const (
	None Relation = iota
	DefaultAction
	Action
	Delete
	Related
	Menu
)

var relationSerializations = map[Relation][]byte{
	None:          []byte(`""`),
	DefaultAction: []byte(`"org.refude.defaultaction"`),
	Action:        []byte(`"org.refude.action"`),
	Delete:        []byte(`"org.refude.delete"`),
	Related:       []byte(`"related"`),
	Menu:          []byte(`"org.refude.menu"`),
}

func (r Relation) MarshalJSON() ([]byte, error) {
	return relationSerializations[r], nil
}


